package srv

import (
	"context"
	"net/http"

	"srv.exe.dev/db/dbgen"
)

// Site selection handlers
func (s *Server) HandleSelectSite(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sites, _ := s.Queries.ListSites(ctx)
	s.render(w, "select_site", map[string]any{"Sites": sites})
}

func (s *Server) HandleSetSite(w http.ResponseWriter, r *http.Request) {
	siteID := r.FormValue("site_id")
	http.SetCookie(w, &http.Cookie{
		Name:     SiteCookieName,
		Value:    siteID,
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Data structures for views
type DashboardData struct {
	Site            *SiteView
	ActiveAlerts    int64
	TotalCameras    int64
	HighRiskZones   int64
	DetectionsToday int64
	Alerts          []AlertView
	Plans           []PlanView
}

type SiteView struct {
	ID          int64
	Name        string
	Location    string
	Description string
}

type AlertView struct {
	ID          int64
	Level       string
	Status      string
	Explanation string
	ZoneName    string
	SentAt      string
}

type PlanView struct {
	ID          int64
	Level       string
	ImagePath   string
	ScaleFactor float64
}

type CameraView struct {
	ID          int64
	Name        string
	StreamURL   string
	PlanID      int64
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
	Polygon   string
	RiskLevel string
	PlanID    int64
	Level     string
	SiteName  string
	IsActive  bool
}

func (s *Server) siteViewFromDB(site *dbgen.Site) *SiteView {
	if site == nil {
		return nil
	}
	return &SiteView{
		ID:          int64(site.ID), // Conversion int32 -> int64
		Name:        site.Name,
		Location:    ptrStr(site.Location),
		Description: ptrStr(site.Description),
	}
}

// Dashboard handler
func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)

	// Count stats for this site
	activeAlerts, _ := s.Queries.CountActiveAlertsBySite(ctx, &siteID)
	totalCameras, _ := s.Queries.CountCamerasBySite(ctx, &siteID)
	highRiskZones, _ := s.Queries.CountHighRiskZonesBySite(ctx, &siteID)
	detections, _ := s.Queries.CountDetectionsTodayBySite(ctx, &siteID)

	// Get alerts for this site
	alerts, _ := s.Queries.ListAlertsBySite(ctx, &siteID)
	var alertViews []AlertView
	for _, a := range alerts {
		alertViews = append(alertViews, AlertView{
			ID:          int64(a.ID), // Conversion int32 -> int64
			Level:       ptrStr(a.AlertLevel),
			Status:      ptrStr(a.Status),
			Explanation: ptrStr(a.Explanation),
			ZoneName:    ptrStr(a.ZoneName),
			SentAt:      a.SentAt.Format("15:04"),
		})
	}

	// Get plans for this site
	plans, _ := s.Queries.ListPlansBySite(ctx, &siteID)
	var planViews []PlanView
	for _, p := range plans {
		planViews = append(planViews, PlanView{
			ID:          int64(p.ID), // Conversion int32 -> int64
			Level:       ptrStr(p.Level),
			ImagePath:   ptrStr(p.ImagePath),
			ScaleFactor: ptrFloat(p.ScaleFactor),
		})
	}

	data := DashboardData{
		Site:            s.siteViewFromDB(site),
		ActiveAlerts:    activeAlerts,
		TotalCameras:    totalCameras,
		HighRiskZones:   highRiskZones,
		DetectionsToday: detections,
		Alerts:          alertViews,
		Plans:           planViews,
	}
	s.render(w, "dashboard", data)
}

// Other page handlers
func (s *Server) HandleCameras(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)
	cameras, _ := s.Queries.ListCamerasBySite(ctx, &siteID)

	var views []CameraView
	for _, c := range cameras {
		views = append(views, CameraView{
			ID:          int64(c.ID), // Conversion int32 -> int64
			Name:        ptrStr(c.Name),
			StreamURL:   ptrStr(c.StreamUrl),
			PlanID:      ptrInt(c.PlanID),
			Level:       ptrStr(c.PlanLevel),
			XPlan:       ptrFloat(c.XPlan),
			YPlan:       ptrFloat(c.YPlan),
			Orientation: ptrFloat(c.Orientation),
			FOV:         ptrFloat(c.Fov),
			IsWebcam:    ptrInt(c.IsWebcam) == 1,
		})
	}
	s.render(w, "cameras", map[string]any{"Site": s.siteViewFromDB(site), "Cameras": views})
}

func (s *Server) HandleAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)
	alerts, _ := s.Queries.ListAlertsBySite(ctx, &siteID)

	var views []AlertView
	for _, a := range alerts {
		views = append(views, AlertView{
			ID:          int64(a.ID), // Conversion int32 -> int64
			Level:       ptrStr(a.AlertLevel),
			Status:      ptrStr(a.Status),
			Explanation: ptrStr(a.Explanation),
			ZoneName:    ptrStr(a.ZoneName),
			SentAt:      a.SentAt.Format("02/01 15:04"),
		})
	}
	s.render(w, "alertes", map[string]any{"Site": s.siteViewFromDB(site), "Alerts": views})
}

func (s *Server) HandleZones(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)
	zones, _ := s.Queries.ListZonesBySite(ctx, &siteID)

	siteName := ""
	if site != nil {
		siteName = site.Name
	}

	var views []ZoneView
	for _, z := range zones {
		views = append(views, ZoneView{
			ID:        int64(z.ID),    // Conversion int32 -> int64
			Name:      ptrStr(z.Name), // <--- Si ptrStr renvoie l'objet NullString,
			Type:      ptrStr(z.Type),
			Polygon:   z.Polygon,
			RiskLevel: ptrStr(z.RiskLevel),
			PlanID:    ptrInt(z.PlanID),
			Level:     ptrStr(z.PlanLevel),
			SiteName:  siteName,
			IsActive:  ptrInt(z.IsActive) == 1,
		})
	}
	s.render(w, "zones", map[string]any{"Site": s.siteViewFromDB(site), "Zones": views})
}

func (s *Server) HandleAnalyses(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	models, _ := s.Queries.ListModels(ctx)
	rules, _ := s.Queries.ListHSERules(ctx)
	s.render(w, "analyses", map[string]any{"Site": s.siteViewFromDB(site), "Models": models, "Rules": rules})
}

func (s *Server) HandleRapports(w http.ResponseWriter, r *http.Request) {
	site := s.getCurrentSite(r)
	s.render(w, "rapports", map[string]any{"Site": s.siteViewFromDB(site)})
}

func (s *Server) HandleValidation(w http.ResponseWriter, r *http.Request) {
	site := s.getCurrentSite(r)
	s.render(w, "validation", map[string]any{
		"Site": s.siteViewFromDB(site),
	})
}

func (s *Server) HandleSuivi(w http.ResponseWriter, r *http.Request) {
	site := s.getCurrentSite(r)
	s.render(w, "suivi", map[string]any{
		"Site": s.siteViewFromDB(site),
	})
}
