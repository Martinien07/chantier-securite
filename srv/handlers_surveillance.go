package srv

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

// ============================================================
// Page HTML — /surveillance
// ============================================================

// HandleSurveillance rend la page de surveillance en direct.
func (s *Server) HandleSurveillance(w http.ResponseWriter, r *http.Request) {
	siteID := s.getCurrentSiteID(r)
	site := s.getCurrentSite(r)

	s.render(w, "surveillance", map[string]any{
		"Site":      s.siteViewFromDB(site),
		"SiteID":    siteID,
		"AIRunning": s.AIStatus(int32(siteID)),
	})
}

// ============================================================
// GET /api/detection/status
// ============================================================

// HandleAPIDetectionStatus retourne si la détection est active pour le site courant.
func (s *Server) HandleAPIDetectionStatus(w http.ResponseWriter, r *http.Request) {
	siteID := s.getCurrentSiteID(r)
	running := s.AIStatus(int32(siteID))
	s.jsonResponse(w, map[string]any{
		"running": running,
		"site_id": siteID,
	})
}

// ============================================================
// GET /api/pending-events
// Retourne les risk_events en attente de validation (status = 'pending').
// Jointure : risk_events → activity_windows → cameras → plans → sites
// ============================================================

type PendingEventView struct {
	ID          int64     `json:"id"`
	CameraID    int64     `json:"camera_id"`
	CameraName  string    `json:"camera_name"`
	ZoneID      int64     `json:"zone_id"`
	ZoneName    string    `json:"zone_name"`
	RuleName    string    `json:"rule_name"`
	RiskScore   float64   `json:"risk_score"`
	RiskLevel   string    `json:"risk_level"`
	Explanation string    `json:"explanation"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Server) HandleAPIPendingEvents(w http.ResponseWriter, r *http.Request) {
	siteID := s.getCurrentSiteID(r)

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT
			re.id,
			COALESCE(aw.camera_id, 0)       AS camera_id,
			COALESCE(c.name, '')             AS camera_name,
			COALESCE(re.zone_id, 0)          AS zone_id,
			COALESCE(z.name, '')             AS zone_name,
			COALESCE(h.name, '')             AS rule_name,
			COALESCE(re.risk_score, 0)       AS risk_score,
			COALESCE(re.risk_level, 'LOW')   AS risk_level,
			COALESCE(re.explanation, '')     AS explanation,
			re.created_at
		FROM risk_events re
		LEFT JOIN activity_windows aw ON aw.id = re.window_id
		LEFT JOIN cameras c           ON c.id  = aw.camera_id
		LEFT JOIN plans   p           ON p.id  = c.plan_id
		LEFT JOIN zones   z           ON z.id  = re.zone_id
		LEFT JOIN hse_rules h         ON h.id  = re.rule_id
		WHERE p.site_id = ?
		  AND re.status = 'pending'
		ORDER BY re.created_at DESC
		LIMIT 50
	`, siteID)
	if err != nil {
		s.jsonResponse(w, []PendingEventView{})
		return
	}
	defer rows.Close()

	var events []PendingEventView
	for rows.Next() {
		var e PendingEventView
		if err := rows.Scan(
			&e.ID, &e.CameraID, &e.CameraName,
			&e.ZoneID, &e.ZoneName,
			&e.RuleName, &e.RiskScore, &e.RiskLevel,
			&e.Explanation, &e.CreatedAt,
		); err != nil {
			continue
		}
		events = append(events, e)
	}

	if events == nil {
		events = []PendingEventView{}
	}
	s.jsonResponse(w, events)
}

// ============================================================
// POST /api/risk-events/{id}/validate
// Crée une alerte dans alerts et passe le risk_event en 'validated'.
// ============================================================

func (s *Server) HandleAPIValidateEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id invalide", http.StatusBadRequest)
		return
	}

	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Récupérer les infos nécessaires pour créer l'alerte
	var cameraID sql.NullInt64
	var riskLevel string
	err = tx.QueryRowContext(r.Context(), `
		SELECT aw.camera_id, COALESCE(re.risk_level, 'LOW')
		FROM risk_events re
		LEFT JOIN activity_windows aw ON aw.id = re.window_id
		WHERE re.id = ? AND re.status = 'pending'
	`, eventID).Scan(&cameraID, &riskLevel)

	if err == sql.ErrNoRows {
		http.Error(w, "événement introuvable ou déjà traité", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Créer l'alerte dans la table alerts
	res, err := tx.ExecContext(r.Context(), `
		INSERT INTO alerts (risk_event_id, camera_id, alert_level, status, sent_at)
		VALUES (?, ?, ?, 'new', NOW())
	`, eventID, cameraID, riskLevel)
	if err != nil {
		http.Error(w, "création alerte: "+err.Error(), http.StatusInternalServerError)
		return
	}
	alertID, _ := res.LastInsertId()

	// Marquer le risk_event comme validé
	_, err = tx.ExecContext(r.Context(),
		`UPDATE risk_events SET status = 'validated' WHERE id = ?`, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]any{
		"message":  "alerte créée",
		"alert_id": alertID,
	})
}
