package srv

import (
	"net/http"
)

// HandleAPICameraStatus — GET /api/cameras/status
// Retourne le statut en ligne/hors ligne de chaque caméra du site.
func (s *Server) HandleAPICameraStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	siteID := s.getCurrentSiteID(r)

	rows, err := s.DB.QueryContext(ctx, `
		SELECT c.id, c.name, c.is_online, c.last_check, c.online_error
		FROM cameras c
		JOIN plans p ON c.plan_id = p.id
		WHERE p.site_id = ?
		ORDER BY c.name
	`, siteID)
	if err != nil {
		s.jsonResponse(w, []any{})
		return
	}
	defer rows.Close()

	type CameraStatus struct {
		CameraID  int64   `json:"camera_id"`
		Name      string  `json:"name"`
		IsOnline  bool    `json:"is_online"`
		LastCheck *string `json:"last_check"`
		ErrorMsg  *string `json:"error_msg"`
	}

	var result []CameraStatus
	for rows.Next() {
		var cs CameraStatus
		var isOnline int
		rows.Scan(&cs.CameraID, &cs.Name, &isOnline, &cs.LastCheck, &cs.ErrorMsg)
		cs.IsOnline = isOnline == 1
		result = append(result, cs)
	}
	if result == nil {
		result = []CameraStatus{}
	}
	s.jsonResponse(w, result)
}

// HandleAPIAIStatus — GET /api/ai/status
// Retourne si le processus IA Python tourne pour ce site.
func (s *Server) HandleAPIAIStatus(w http.ResponseWriter, r *http.Request) {
	siteID := s.getCurrentSiteID(r)
	running := s.AIStatus(int32(siteID))
	s.jsonResponse(w, map[string]any{
		"running": running,
		"site_id": siteID,
	})
}
