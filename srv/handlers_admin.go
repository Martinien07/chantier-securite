package srv

import (
	"context"
	"net/http"
)

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)

	plans, _ := s.Queries.ListPlansBySite(ctx, &siteID)
	cameras, _ := s.Queries.CountCamerasBySite(ctx, &siteID)
	zones, _ := s.Queries.ListZonesBySite(ctx, &siteID)
	users, _ := s.Queries.ListUsers(ctx)
	rules, _ := s.Queries.ListHSERules(ctx)
	models, _ := s.Queries.ListModels(ctx)

	s.render(w, "admin", map[string]any{
		"Site":        s.siteViewFromDB(site),
		"PlanCount":   len(plans),
		"CameraCount": cameras,
		"ZoneCount":   len(zones),
		"UserCount":   len(users),
		"RuleCount":   len(rules),
		"ModelCount":  len(models),
	})
}

func (s *Server) HandleAdminPlans(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)
	plans, _ := s.Queries.ListPlansBySite(ctx, &siteID)

	var views []PlanView
	for _, p := range plans {
		views = append(views, PlanView{
			ID:          p.ID,
			Level:       ptrStr(p.Level),
			ImagePath:   ptrStr(p.ImagePath),
			ScaleFactor: ptrFloat(p.ScaleFactor),
		})
	}
	s.render(w, "admin_plans", map[string]any{"Site": s.siteViewFromDB(site), "Plans": views})
}

func (s *Server) HandleAdminCameras(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)
	cameras, _ := s.Queries.ListCamerasBySite(ctx, &siteID)
	plans, _ := s.Queries.ListPlansBySite(ctx, &siteID)

	var views []CameraView
	for _, c := range cameras {
		views = append(views, CameraView{
			ID:          c.ID,
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

	var planViews []PlanView
	for _, p := range plans {
		planViews = append(planViews, PlanView{
			ID:        p.ID,
			Level:     ptrStr(p.Level),
			ImagePath: ptrStr(p.ImagePath),
		})
	}
	s.render(w, "admin_cameras", map[string]any{"Site": s.siteViewFromDB(site), "Cameras": views, "Plans": planViews})
}

func (s *Server) HandleAdminZones(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	siteID := s.getCurrentSiteID(r)
	zones, _ := s.Queries.ListZonesBySite(ctx, &siteID)
	plans, _ := s.Queries.ListPlansBySite(ctx, &siteID)

	var views []ZoneView
	for _, z := range zones {
		views = append(views, ZoneView{
			ID:        z.ID,
			Name:      ptrStr(z.Name),
			Type:      ptrStr(z.Type),
			Polygon:   z.Polygon,
			RiskLevel: ptrStr(z.RiskLevel),
			PlanID:    ptrInt(z.PlanID),
			Level:     ptrStr(z.PlanLevel),
			IsActive:  ptrInt(z.IsActive) == 1,
		})
	}

	var planViews []PlanView
	for _, p := range plans {
		planViews = append(planViews, PlanView{
			ID:    p.ID,
			Level: ptrStr(p.Level),
		})
	}
	s.render(w, "admin_zones", map[string]any{"Site": s.siteViewFromDB(site), "Zones": views, "Plans": planViews})
}

func (s *Server) HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	users, _ := s.Queries.ListUsers(ctx)
	roles, _ := s.Queries.ListRoles(ctx)
	s.render(w, "admin_users", map[string]any{"Site": s.siteViewFromDB(site), "Users": users, "Roles": roles})
}

func (s *Server) HandleAdminRules(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	rules, _ := s.Queries.ListHSERules(ctx)
	s.render(w, "admin_rules", map[string]any{"Site": s.siteViewFromDB(site), "Rules": rules})
}

func (s *Server) HandleAdminModels(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	site := s.getCurrentSite(r)
	models, _ := s.Queries.ListModels(ctx)
	s.render(w, "admin_models", map[string]any{"Site": s.siteViewFromDB(site), "Models": models})
}
