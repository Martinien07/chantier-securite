// srv/handlers_riskevent.go
package srv

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"srv.exe.dev/db/dbgen"
)

func (s *Server) HandleAPIPendingRiskEvents(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)

	// Fallback : si pas de site dans le cookie, prendre tous les pending
	var events []dbgen.RiskEventPending
	var err error
	if siteID == 0 {
		// Pas de site sélectionné — retourner liste vide
		s.jsonResponse(w, []dbgen.RiskEventPending{})
		return
	}

	events, err = s.Queries.ListPendingRiskEventsBySite(ctx, &siteID)
	if err != nil {
		// En cas d'erreur JOIN (plan manquant), requête sans filtre site
		rows, err2 := s.DB.QueryContext(ctx,
			`SELECT re.id, re.camera_id, re.risk_level, re.risk_score,
			        re.risk_messages, re.explanation,
			        re.window_start, re.window_end, re.created_at,
			        c.name as camera_name, '' as site_name
			 FROM risk_events re
			 LEFT JOIN cameras c ON re.camera_id = c.id
			 WHERE re.status = 'pending'
			 ORDER BY re.created_at DESC LIMIT 100`)
		if err2 != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var e dbgen.RiskEventPending
			rows.Scan(&e.ID, &e.CameraID, &e.RiskLevel, &e.RiskScore,
				&e.RiskMessages, &e.Explanation,
				&e.WindowStart, &e.WindowEnd, &e.CreatedAt,
				&e.CameraName, &e.SiteName)
			events = append(events, e)
		}
	}
	if events == nil {
		events = []dbgen.RiskEventPending{}
	}
	s.jsonResponse(w, events)
}

func (s *Server) HandleAPIAcceptRiskEvent(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id invalide", 400)
		return
	}

	var req struct {
		Comment string `json:"comment"`
		Action  string `json:"action"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Comment == "" {
		http.Error(w, "commentaire obligatoire", 400)
		return
	}
	if req.Action == "" {
		req.Action = "accepted"
	}

	if err := s.Queries.AcceptRiskEvent(ctx, id, req.Comment, req.Action); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.jsonResponse(w, map[string]string{"status": req.Action})
}

func (s *Server) HandleAPIRejectRiskEvent(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id invalide", 400)
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Comment == "" {
		http.Error(w, "commentaire obligatoire", 400)
		return
	}

	if err := s.Queries.RejectRiskEvent(ctx, id, req.Comment); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.jsonResponse(w, map[string]string{"status": "rejected"})
}

func (s *Server) HandleAPIPendingCount(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)
	count, _ := s.Queries.CountPendingRiskEventsBySite(ctx, &siteID)
	s.jsonResponse(w, map[string]int64{"count": count})
}
