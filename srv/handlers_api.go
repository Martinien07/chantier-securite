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
	alerts, _ := s.Queries.ListAlertsBySite(ctx, &siteID)
	s.jsonResponse(w, alerts)
}

func (s *Server) HandleAPICameras(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)
	cameras, _ := s.Queries.ListCamerasBySite(ctx, &siteID)
	s.jsonResponse(w, cameras)
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
	s.jsonResponse(w, zones)
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

	s.jsonResponse(w, map[string]any{
		"plan":    plan,
		"cameras": cameras,
		"zones":   zones,
	})
}

func (s *Server) HandleAPIRecentDetections(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID := s.getCurrentSiteID(r)
	detections, _ := s.Queries.GetRecentDetectionsBySite(ctx, dbgen.GetRecentDetectionsBySiteParams{
		SiteID: &siteID,
		Limit:  100,
	})
	s.jsonResponse(w, detections)
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
	s.Queries.UpdatePlan(ctx, dbgen.UpdatePlanParams{
		ID:          planID,
		Level:       plan.Level,
		ImagePath:   &imagePath,
		ScaleFactor: plan.ScaleFactor,
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
