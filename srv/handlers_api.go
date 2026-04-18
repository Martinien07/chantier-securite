package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"srv.exe.dev/db/dbgen"
)

func (s *Server) HandleAPIStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)

	activeAlerts, _ := s.Queries.CountActiveAlertsBySite(ctx, &siteID)
	cameras, _ := s.Queries.CountCamerasBySite(ctx, &siteID)
	highRisk, _ := s.Queries.CountHighRiskZonesBySite(ctx, &siteID)
	detections, _ := s.Queries.CountDetectionsTodayBySite(ctx, &siteID)

	s.jsonResponse(w, map[string]int64{
		"active_alerts":    activeAlerts,
		"total_cameras":    cameras,
		"high_risk_zones":  highRisk,
		"detections_today": detections,
	})
}

func (s *Server) HandleAPIAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)

	type AlertJSON struct {
		ID          int64  `json:"id"`
		AlertLevel  string `json:"alert_level"`
		Status      string `json:"status"`
		Explanation string `json:"explanation"`
		ZoneName    string `json:"zone_name"`
		CameraName  string `json:"camera_name"`
		SentAt      string `json:"sent_at"`
	}

	alerts, _ := s.Queries.ListAlertsBySite(ctx, &siteID)
	result := []AlertJSON{}
	for _, a := range alerts {
		result = append(result, AlertJSON{
			ID:          int64(a.ID),
			AlertLevel:  ptrStr(a.AlertLevel),
			Status:      ptrStr(a.Status),
			Explanation: ptrStr(a.Explanation),
			ZoneName:    ptrStr(a.ZoneName),
			CameraName:  ptrStr(a.CameraName),
			SentAt:      a.SentAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	s.jsonResponse(w, result)
}

func (s *Server) HandleAPICameras(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)
	cameras, _ := s.Queries.ListCamerasBySite(ctx, &siteID)

	type CameraJSON struct {
		ID          int64   `json:"id"`
		Name        string  `json:"name"`
		StreamURL   string  `json:"stream_url"`
		XPlan       float64 `json:"x_plan"`
		YPlan       float64 `json:"y_plan"`
		Orientation float64 `json:"orientation"`
		FOV         float64 `json:"fov"`
		IsWebcam    int64   `json:"is_webcam"`
		PlanID      int64   `json:"plan_id"`
		PlanLevel   string  `json:"plan_level"`
	}
	out := make([]CameraJSON, 0, len(cameras))
	for _, c := range cameras {
		out = append(out, CameraJSON{
			ID:          int64(c.ID),
			Name:        ptrStr(c.Name),
			StreamURL:   ptrStr(c.StreamUrl),
			XPlan:       ptrFloat(c.XPlan),
			YPlan:       ptrFloat(c.YPlan),
			Orientation: ptrFloat(c.Orientation),
			FOV:         ptrFloat(c.Fov),
			IsWebcam:    ptrInt(c.IsWebcam),
			PlanID:      ptrInt(c.PlanID),
			PlanLevel:   ptrStr(c.PlanLevel),
		})
	}
	s.jsonResponse(w, out)
}

func (s *Server) HandleAPISites(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sites, _ := s.Queries.ListSites(ctx)
	s.jsonResponse(w, sites)
}

func (s *Server) HandleAPISearchSites(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	q := "%" + r.URL.Query().Get("q") + "%"
	sites, _ := s.Queries.SearchSites(ctx, dbgen.SearchSitesParams{Name: q, Location: &q})
	s.jsonResponse(w, sites)
}

func (s *Server) HandleAPIZones(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)
	zones, _ := s.Queries.ListZonesBySite(ctx, &siteID)

	type ZoneJSON struct {
		ID        int64   `json:"id"`
		Name      string  `json:"name"`
		Type      string  `json:"type"`
		Polygon   string  `json:"polygon"`
		RiskLevel string  `json:"risk_level"`
		IsActive  int64   `json:"is_active"`
		PlanID    int64   `json:"plan_id"`
		PlanLevel string  `json:"plan_level"`
	}
	out := make([]ZoneJSON, 0, len(zones))
	for _, z := range zones {
		out = append(out, ZoneJSON{
			ID:        int64(z.ID),
			Name:      ptrStr(z.Name),
			Type:      ptrStr(z.Type),
			Polygon:   z.Polygon,
			RiskLevel: ptrStr(z.RiskLevel),
			IsActive:  ptrInt(z.IsActive),
			PlanID:    ptrInt(z.PlanID),
			PlanLevel: ptrStr(z.PlanLevel),
		})
	}
	s.jsonResponse(w, out)
}

func (s *Server) HandleAPIPlans(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)
	plans, _ := s.Queries.ListPlansBySite(ctx, &siteID)
	s.jsonResponse(w, plans)
}

// Get plan data with cameras and zones
func (s *Server) HandleAPIPlanData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	planID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	plan, err := s.Queries.GetPlan(ctx, planID)
	if err != nil {
		http.Error(w, "Plan not found", 404)
		return
	}

	cameras, _ := s.Queries.ListCamerasByPlan(ctx, &planID)
	zones, _ := s.Queries.ListZonesByPlan(ctx, &planID)

	// Sérialisation propre — tous les *string/*float64/*int64 Go
	// sont convertis en valeurs JSON simples pour le frontend.
	type PlanJSON struct {
		ID          int64   `json:"id"`
		Level       string  `json:"level"`
		ImagePath   string  `json:"image_path"`
		ScaleFactor float64 `json:"scale_factor"`
	}
	type CameraJSON struct {
		ID          int64   `json:"id"`
		Name        string  `json:"name"`
		StreamURL   string  `json:"stream_url"`
		XPlan       float64 `json:"x_plan"`
		YPlan       float64 `json:"y_plan"`
		Orientation float64 `json:"orientation"`
		FOV         float64 `json:"fov"`
		IsWebcam    int64   `json:"is_webcam"`
		PlanID      int64   `json:"plan_id"`
	}
	type ZoneJSON struct {
		ID        int64   `json:"id"`
		Name      string  `json:"name"`
		Type      string  `json:"type"`
		Polygon   string  `json:"polygon"`
		RiskLevel string  `json:"risk_level"`
		IsActive  int64   `json:"is_active"`
		PlanID    int64   `json:"plan_id"`
	}

	planOut := PlanJSON{
		ID:          int64(plan.ID),
		Level:       ptrStr(plan.Level),
		ImagePath:   ptrStr(plan.ImagePath),
		ScaleFactor: ptrFloat(plan.ScaleFactor),
	}

	camOut := make([]CameraJSON, 0, len(cameras))
	for _, c := range cameras {
		camOut = append(camOut, CameraJSON{
			ID:          int64(c.ID),
			Name:        ptrStr(c.Name),
			StreamURL:   ptrStr(c.StreamUrl),
			XPlan:       ptrFloat(c.XPlan),
			YPlan:       ptrFloat(c.YPlan),
			Orientation: ptrFloat(c.Orientation),
			FOV:         ptrFloat(c.Fov),
			IsWebcam:    ptrInt(c.IsWebcam),
			PlanID:      ptrInt(c.PlanID),
		})
	}

	zoneOut := make([]ZoneJSON, 0, len(zones))
	for _, z := range zones {
		zoneOut = append(zoneOut, ZoneJSON{
			ID:        int64(z.ID),
			Name:      ptrStr(z.Name),
			Type:      ptrStr(z.Type),
			Polygon:   z.Polygon,
			RiskLevel: ptrStr(z.RiskLevel),
			IsActive:  ptrInt(z.IsActive),
			PlanID:    ptrInt(z.PlanID),
		})
	}

	s.jsonResponse(w, map[string]any{
		"plan":    planOut,
		"cameras": camOut,
		"zones":   zoneOut,
	})
}

func (s *Server) HandleAPIRecentDetections(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)

	type DetectionJSON struct {
		ID          int64   `json:"id"`
		CameraID    int64   `json:"camera_id"`
		Timestamp   string  `json:"timestamp"`
		ObjectClass string  `json:"object_class"`
		Confidence  float64 `json:"confidence"`
		TrackID     int64   `json:"track_id"`
	}

	// Requête directe filtrée sur aujourd'hui
	rows, err := s.DB.QueryContext(ctx, `
		SELECT d.id, d.camera_id, d.timestamp, d.object_class,
		       d.confidence, d.track_id
		FROM detections d
		JOIN cameras c  ON d.camera_id = c.id
		JOIN plans p    ON c.plan_id   = p.id
		WHERE p.site_id = ?
		  AND DATE(d.timestamp) = CURDATE()
		ORDER BY d.timestamp DESC
		LIMIT 500
	`, siteID)
	if err != nil {
		s.jsonResponse(w, []DetectionJSON{})
		return
	}
	defer rows.Close()

	var result []DetectionJSON
	for rows.Next() {
		var d DetectionJSON
		var ts []byte
		rows.Scan(&d.ID, &d.CameraID, &ts, &d.ObjectClass, &d.Confidence, &d.TrackID)
		d.Timestamp = string(ts)
		result = append(result, d)
	}
	if result == nil {
		result = []DetectionJSON{}
	}
	s.jsonResponse(w, result)
}

func (s *Server) HandleAPIAckAlert(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	now := time.Now()
	s.Queries.AcknowledgeAlert(ctx, dbgen.AcknowledgeAlertParams{
		ID:             id,
		AcknowledgedAt: &now,
	})
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) HandleAPICloseAlert(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	s.Queries.CloseAlert(ctx, id)
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) HandleAPIResolveAlert(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	// Résoudre = fermer avec statut 'closed' + marquer le risk_event associé comme résolu
	_, err := s.DB.ExecContext(ctx,
		`UPDATE alerts SET status = 'closed', acknowledged_at = NOW() WHERE id = ?`, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Marquer aussi le risk_event comme résolu
	s.DB.ExecContext(ctx,
		`UPDATE risk_events re
		 JOIN alerts a ON a.risk_event_id = re.id
		 SET re.status = 'resolved', re.resolved_at = NOW()
		 WHERE a.id = ?`, id)
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

// Plan management
func (s *Server) HandleAPICreatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)

	var req struct {
		Level       string  `json:"level"`
		ScaleFactor float64 `json:"scale_factor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	plan, err := s.Queries.CreatePlan(ctx, dbgen.CreatePlanParams{
		SiteID:      &siteID,
		Level:       &req.Level,
		ScaleFactor: &req.ScaleFactor,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.jsonResponse(w, plan)
}

func (s *Server) HandleAPIUploadPlanImage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	planID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	// Max 10MB
	r.ParseMultipartForm(10 << 20)

	_, file, _, _ := runtime.Caller(0)
	uploadDir := filepath.Join(filepath.Dir(file), "static", "uploads", "plans")

	filename, err := saveUploadedFile(r, "image", uploadDir)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	imagePath := fmt.Sprintf("/static/uploads/plans/%s", filename)

	plan, _ := s.Queries.GetPlan(ctx, planID)
	// --- DANS handlers_api.go ---

	// Dans la fonction HandleAPIUploadPlanImage :
	s.Queries.UpdatePlan(ctx, dbgen.UpdatePlanParams{
		ID:          planID,
		Level:       strPtr(ptrStr(plan.Level)), // Correction ici
		ImagePath:   &imagePath,
		ScaleFactor: floatPtr(ptrFloat(plan.ScaleFactor)), // Correction ici

	})

	s.jsonResponse(w, map[string]string{"image_path": imagePath})
}

func (s *Server) HandleAPIDeletePlan(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	s.Queries.DeletePlan(ctx, id)
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

// Camera management
func (s *Server) HandleAPICreateCamera(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req struct {
		PlanID      int64   `json:"plan_id"`
		Name        string  `json:"name"`
		StreamURL   string  `json:"stream_url"`
		XPlan       float64 `json:"x_plan"`
		YPlan       float64 `json:"y_plan"`
		Orientation float64 `json:"orientation"`
		FOV         float64 `json:"fov"`
		IsWebcam    bool    `json:"is_webcam"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	isWebcam := int64(0)
	if req.IsWebcam {
		isWebcam = 1
	}
	cam, err := s.Queries.CreateCamera(ctx, dbgen.CreateCameraParams{
		PlanID:      &req.PlanID,
		Name:        &req.Name,
		StreamUrl:   &req.StreamURL,
		XPlan:       &req.XPlan,
		YPlan:       &req.YPlan,
		Orientation: &req.Orientation,
		Fov:         &req.FOV,
		IsWebcam:    &isWebcam,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.jsonResponse(w, cam)
}

func (s *Server) HandleAPIUpdateCamera(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Name        string  `json:"name"`
		StreamURL   string  `json:"stream_url"`
		XPlan       float64 `json:"x_plan"`
		YPlan       float64 `json:"y_plan"`
		Orientation float64 `json:"orientation"`
		FOV         float64 `json:"fov"`
		IsWebcam    bool    `json:"is_webcam"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	isWebcam := int64(0)
	if req.IsWebcam {
		isWebcam = 1
	}
	s.Queries.UpdateCamera(ctx, dbgen.UpdateCameraParams{
		ID:          id,
		Name:        &req.Name,
		StreamUrl:   &req.StreamURL,
		XPlan:       &req.XPlan,
		YPlan:       &req.YPlan,
		Orientation: &req.Orientation,
		Fov:         &req.FOV,
		IsWebcam:    &isWebcam,
	})
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) HandleAPIDeleteCamera(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	s.Queries.DeleteCamera(ctx, id)
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

// Zone management
func (s *Server) HandleAPICreateZone(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req struct {
		PlanID    int64  `json:"plan_id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		Polygon   string `json:"polygon"`
		RiskLevel string `json:"risk_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	zone, err := s.Queries.CreateZone(ctx, dbgen.CreateZoneParams{
		PlanID:    &req.PlanID,
		Name:      &req.Name,
		Type:      &req.Type,
		Polygon:   req.Polygon,
		RiskLevel: &req.RiskLevel,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.jsonResponse(w, zone)
}

func (s *Server) HandleAPIUpdateZone(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Name      string `json:"name"`
		Type      string `json:"type"`
		Polygon   string `json:"polygon"`
		RiskLevel string `json:"risk_level"`
		IsActive  bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	isActive := int64(0)
	if req.IsActive {
		isActive = 1
	}
	s.Queries.UpdateZone(ctx, dbgen.UpdateZoneParams{
		ID:        id,
		Name:      &req.Name,
		Type:      &req.Type,
		Polygon:   req.Polygon,
		RiskLevel: &req.RiskLevel,
		IsActive:  &isActive,
	})
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) HandleAPIDeleteZone(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	s.Queries.DeleteZone(ctx, id)
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) HandleAPIGetZone(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	zone, err := s.Queries.GetZone(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s.jsonResponse(w, zone)
}

// HandleAPIActiveAlerts returns alerts with camera_id for plan overlay
func (s *Server) HandleAPIActiveAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)
	alerts, err := s.Queries.ListAlertsBySite(ctx, &siteID)
	if err != nil {
		s.jsonResponse(w, []struct{}{})
		return
	}

	var result []map[string]any
	for _, a := range alerts {
		if ptrStr(a.Status) != "new" && ptrStr(a.Status) != "acknowledged" {
			continue
		}
		result = append(result, map[string]any{
			"id":          a.ID,
			"alert_level": ptrStr(a.AlertLevel),
			"status":      ptrStr(a.Status),
			"camera_id":   ptrInt(a.CameraID),
			"camera_name": ptrStr(a.CameraName),
			"explanation": ptrStr(a.Explanation),
		})
	}

	if result == nil {
		result = []map[string]any{}
	}
	s.jsonResponse(w, result)
}

func (s *Server) HandleAPIUpdateAlert(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	var req struct {
		Level  string `json:"level"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Update alert
	_, err := s.DB.ExecContext(ctx,
		"UPDATE alerts SET alert_level = ?, status = ? WHERE id = ?",
		req.Level, req.Status, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.jsonResponse(w, map[string]string{"status": "ok"})
}

// HandleAPIPendingLocations — GET /api/pending-locations
// Retourne les risk_events pending avec les dernières positions détectées
// et la matrice homographique pour projection sur le plan.
func (s *Server) HandleAPIPendingLocations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	siteID := s.getCurrentSiteID(r)

	type Location struct {
		RiskEventID int64       `json:"risk_event_id"`
		CameraID    int64       `json:"camera_id"`
		CameraName  string      `json:"camera_name"`
		RiskLevel   string      `json:"risk_level"`
		RiskMsg     string      `json:"risk_messages"`
		ZoneName    string      `json:"zone_name"`
		PlanLevel   string      `json:"plan_level"`
		PlanID      int64       `json:"plan_id"`
		SiteName    string      `json:"site_name"`
		Status      string      `json:"status"`
		// Pied de la bbox en pixels (centre_x, bas_y)
		FootX       float64     `json:"foot_x"`
		FootY       float64     `json:"foot_y"`
		// Matrice homographique 3x3
		Homography  interface{} `json:"homography"`
		// Coordonnées projetées sur le plan (calculées si H disponible)
		PlanX       *float64    `json:"plan_x"`
		PlanY       *float64    `json:"plan_y"`
		CreatedAt   string      `json:"created_at"`
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT
			re.id,
			re.camera_id,
			COALESCE(c.name, ''),
			re.risk_level,
			re.risk_messages,
			COALESCE(z.name, ''),
			COALESCE(p.level, ''),
			p.id as plan_id,
			COALESCE(s.name, ''),
			re.status,
			-- Dernière détection pour cette caméra (pied de la bbox)
			COALESCE(d.bbox_x + d.bbox_w/2, 0) as foot_x,
			COALESCE(d.bbox_y + d.bbox_h,   0) as foot_y,
			-- Matrice homographique
			COALESCE(cc.homography, '[]'),
			re.created_at
		FROM risk_events re
		JOIN cameras c      ON re.camera_id = c.id
		JOIN plans p        ON c.plan_id    = p.id
		JOIN sites s        ON p.site_id    = s.id
		LEFT JOIN zones z   ON re.zone_id   = z.id
		LEFT JOIN camera_calibrations cc ON cc.camera_id = c.id AND cc.plan_id = p.id AND cc.is_active = 1
		LEFT JOIN detections d ON d.camera_id = re.camera_id
			AND d.timestamp = (
				SELECT MAX(d2.timestamp) FROM detections d2
				WHERE d2.camera_id = re.camera_id
				  AND d2.timestamp <= re.created_at
			)
		WHERE re.status IN ('pending', 'correcting', 'accepted')
		  AND p.site_id = ?
		ORDER BY re.created_at DESC
	`, siteID)

	if err != nil {
		s.jsonResponse(w, []Location{})
		return
	}
	defer rows.Close()

	var result []Location
	for rows.Next() {
		var loc Location
		var homographyJSON string
		rows.Scan(
			&loc.RiskEventID, &loc.CameraID, &loc.CameraName,
			&loc.RiskLevel, &loc.RiskMsg,
			&loc.ZoneName, &loc.PlanLevel, &loc.PlanID, &loc.SiteName,
			&loc.Status,
			&loc.FootX, &loc.FootY,
			&homographyJSON, &loc.CreatedAt,
		)

		// Parser la matrice H
		var H [3][3]float64
		if homographyJSON != "" && homographyJSON != "[]" {
			if err := parseHomography(homographyJSON, &H); err == nil {
				// Appliquer H : [x', y', w'] = H × [px, py, 1]
				xp := H[0][0]*loc.FootX + H[0][1]*loc.FootY + H[0][2]
				yp := H[1][0]*loc.FootX + H[1][1]*loc.FootY + H[1][2]
				wp := H[2][0]*loc.FootX + H[2][1]*loc.FootY + H[2][2]
				if wp != 0 {
					px := xp / wp
					py := yp / wp
					loc.PlanX = &px
					loc.PlanY = &py
				}
			}
		}

		// Stocker H raw pour le JS (au cas où on veut recalculer côté client)
		loc.Homography = json.RawMessage(homographyJSON)
		result = append(result, loc)
	}

	if result == nil {
		result = []Location{}
	}
	s.jsonResponse(w, result)
}

func parseHomography(jsonStr string, H *[3][3]float64) error {
	var raw [][]float64
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return err
	}
	if len(raw) != 3 {
		return fmt.Errorf("invalid homography size")
	}
	for i := 0; i < 3; i++ {
		if len(raw[i]) != 3 {
			return fmt.Errorf("invalid row size")
		}
		for j := 0; j < 3; j++ {
			H[i][j] = raw[i][j]
		}
	}
	return nil
}

// HandleAPIRevertToPending — POST /api/risk-events/{id}/revert-to-pending
// Remet un risk_event correcting/accepted en pending après timeout de 10min.
func (s *Server) HandleAPIRevertToPending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
		req.Comment = "Retour automatique — non résolu après 10 minutes"
	}
	_, err = s.DB.ExecContext(ctx,
		`UPDATE risk_events SET status = 'pending', operator_comment = ? WHERE id = ? AND status IN ('correcting','accepted')`,
		req.Comment, id,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.jsonResponse(w, map[string]string{"status": "reverted"})
}
