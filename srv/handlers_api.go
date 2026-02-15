package srv

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"srv.exe.dev/db/dbgen"
)

func (s *Server) HandleAPIStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	activeAlerts, _ := s.Queries.CountActiveAlerts(ctx)
	cameras, _ := s.Queries.CountActiveCameras(ctx)
	highRisk, _ := s.Queries.CountHighRiskZones(ctx)
	detections, _ := s.Queries.CountDetectionsToday(ctx)

	s.jsonResponse(w, map[string]int64{
		"active_alerts":    activeAlerts,
		"total_cameras":    cameras,
		"high_risk_zones":  highRisk,
		"detections_today": detections,
	})
}

func (s *Server) HandleAPIAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	alerts, _ := s.Queries.ListAlerts(ctx, 100)
	s.jsonResponse(w, alerts)
}

func (s *Server) HandleAPICameras(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cameras, _ := s.Queries.ListCameras(ctx)
	s.jsonResponse(w, cameras)
}

func (s *Server) HandleAPISites(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sites, _ := s.Queries.ListSites(ctx)
	s.jsonResponse(w, sites)
}

func (s *Server) HandleAPIZones(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	zones, _ := s.Queries.ListZones(ctx)
	s.jsonResponse(w, zones)
}

func (s *Server) HandleAPIPlans(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	siteID, _ := strconv.ParseInt(r.PathValue("siteId"), 10, 64)
	plans, _ := s.Queries.ListPlansBySite(ctx, &siteID)
	s.jsonResponse(w, plans)
}

func (s *Server) HandleAPIRecentDetections(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	detections, _ := s.Queries.GetRecentDetections(ctx, 100)
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
		Name      string `json:"name"`
		StreamURL string `json:"stream_url"`
		IsWebcam  bool   `json:"is_webcam"`
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
		ID:        id,
		Name:      &req.Name,
		StreamUrl: &req.StreamURL,
		IsWebcam:  &isWebcam,
	})
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) HandleAPIDeleteCamera(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	s.Queries.DeleteCamera(ctx, id)
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) HandleAPICreateSite(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req struct {
		Name        string `json:"name"`
		Location    string `json:"location"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	site, err := s.Queries.CreateSite(ctx, dbgen.CreateSiteParams{
		Name:        req.Name,
		Location:    &req.Location,
		Description: &req.Description,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.jsonResponse(w, site)
}

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
