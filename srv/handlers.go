package srv

import (
	"context"
	"net/http"
)

type DashboardData struct {
	ActiveAlerts    int64
	TotalCameras    int64
	HighRiskZones   int64
	DetectionsToday int64
	Alerts          []AlertView
	Sites           []SiteView
}

type AlertView struct {
	ID          int64
	Level       string
	Status      string
	Explanation string
	ZoneName    string
	SentAt      string
}

type SiteView struct {
	ID          int64
	Name        string
	Location    string
	Description string
}

type CameraView struct {
	ID          int64
	Name        string
	StreamURL   string
	PlanID      int64
	SiteName    string
	Level       string
	XPlan       float64
	YPlan       float64
	Orientation float64
	FOV         float64
	IsWebcam    bool
}

type ZoneView struct {
	ID        int64
	Name      string
	Type      string
	RiskLevel string
	PlanID    int64
	Level     string
	SiteName  string
	IsActive  bool
}

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx := context.Background()

	activeAlerts, _ := s.Queries.CountActiveAlerts(ctx)
	totalCameras, _ := s.Queries.CountActiveCameras(ctx)
	highRiskZones, _ := s.Queries.CountHighRiskZones(ctx)
	detections, _ := s.Queries.CountDetectionsToday(ctx)

	alerts, _ := s.Queries.ListAlerts(ctx, 10)
	sites, _ := s.Queries.ListSites(ctx)

	var alertViews []AlertView
	for _, a := range alerts {
		alertViews = append(alertViews, AlertView{
			ID:          a.ID,
			Level:       ptrStr(a.AlertLevel),
			Status:      ptrStr(a.Status),
			Explanation: ptrStr(a.Explanation),
			ZoneName:    ptrStr(a.ZoneName),
			SentAt:      a.SentAt.Format("15:04"),
		})
	}

	var siteViews []SiteView
	for _, site := range sites {
		siteViews = append(siteViews, SiteView{
			ID:          site.ID,
			Name:        site.Name,
			Location:    ptrStr(site.Location),
			Description: ptrStr(site.Description),
		})
	}

	data := DashboardData{
		ActiveAlerts:    activeAlerts,
		TotalCameras:    totalCameras,
		HighRiskZones:   highRiskZones,
		DetectionsToday: detections,
		Alerts:          alertViews,
		Sites:           siteViews,
	}
	s.render(w, "dashboard", data)
}

func (s *Server) HandleCameras(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cameras, _ := s.Queries.ListCameras(ctx)

	var views []CameraView
	for _, c := range cameras {
		views = append(views, CameraView{
			ID:          c.ID,
			Name:        ptrStr(c.Name),
			StreamURL:   ptrStr(c.StreamUrl),
			PlanID:      ptrInt(c.PlanID),
			SiteName:    ptrStr(c.SiteName),
			Level:       ptrStr(c.PlanLevel),
			XPlan:       ptrFloat(c.XPlan),
			YPlan:       ptrFloat(c.YPlan),
			Orientation: ptrFloat(c.Orientation),
			FOV:         ptrFloat(c.Fov),
			IsWebcam:    ptrInt(c.IsWebcam) == 1,
		})
	}
	s.render(w, "cameras", map[string]any{"Cameras": views})
}

func (s *Server) HandleAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	alerts, _ := s.Queries.ListAlerts(ctx, 50)

	var views []AlertView
	for _, a := range alerts {
		views = append(views, AlertView{
			ID:          a.ID,
			Level:       ptrStr(a.AlertLevel),
			Status:      ptrStr(a.Status),
			Explanation: ptrStr(a.Explanation),
			ZoneName:    ptrStr(a.ZoneName),
			SentAt:      a.SentAt.Format("02/01 15:04"),
		})
	}
	s.render(w, "alertes", map[string]any{"Alerts": views})
}

func (s *Server) HandleZones(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	zones, _ := s.Queries.ListZones(ctx)

	var views []ZoneView
	for _, z := range zones {
		views = append(views, ZoneView{
			ID:        z.ID,
			Name:      ptrStr(z.Name),
			Type:      ptrStr(z.Type),
			RiskLevel: ptrStr(z.RiskLevel),
			PlanID:    ptrInt(z.PlanID),
			Level:     ptrStr(z.PlanLevel),
			SiteName:  ptrStr(z.SiteName),
			IsActive:  ptrInt(z.IsActive) == 1,
		})
	}
	s.render(w, "zones", map[string]any{"Zones": views})
}

func (s *Server) HandleAnalyses(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	models, _ := s.Queries.ListModels(ctx)
	rules, _ := s.Queries.ListHSERules(ctx)
	s.render(w, "analyses", map[string]any{"Models": models, "Rules": rules})
}

func (s *Server) HandleRapports(w http.ResponseWriter, r *http.Request) {
	s.render(w, "rapports", nil)
}
