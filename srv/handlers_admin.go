package srv

import (
	"context"
	"net/http"
)

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sites, _ := s.Queries.ListSites(ctx)
	cameras, _ := s.Queries.CountActiveCameras(ctx)
	users, _ := s.Queries.ListUsers(ctx)
	rules, _ := s.Queries.ListHSERules(ctx)
	models, _ := s.Queries.ListModels(ctx)

	s.render(w, "admin", map[string]any{
		"SiteCount":   len(sites),
		"CameraCount": cameras,
		"UserCount":   len(users),
		"RuleCount":   len(rules),
		"ModelCount":  len(models),
	})
}

func (s *Server) HandleAdminSites(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sites, _ := s.Queries.ListSites(ctx)
	s.render(w, "admin_sites", map[string]any{"Sites": sites})
}

func (s *Server) HandleAdminCameras(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cameras, _ := s.Queries.ListCameras(ctx)
	sites, _ := s.Queries.ListSites(ctx)

	var views []CameraView
	for _, c := range cameras {
		views = append(views, CameraView{
			ID:        c.ID,
			Name:      ptrStr(c.Name),
			StreamURL: ptrStr(c.StreamUrl),
			PlanID:    ptrInt(c.PlanID),
			SiteName:  ptrStr(c.SiteName),
			Level:     ptrStr(c.PlanLevel),
			IsWebcam:  ptrInt(c.IsWebcam) == 1,
		})
	}
	s.render(w, "admin_cameras", map[string]any{"Cameras": views, "Sites": sites})
}

func (s *Server) HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	users, _ := s.Queries.ListUsers(ctx)
	roles, _ := s.Queries.ListRoles(ctx)
	s.render(w, "admin_users", map[string]any{"Users": users, "Roles": roles})
}

func (s *Server) HandleAdminRules(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	rules, _ := s.Queries.ListHSERules(ctx)
	s.render(w, "admin_rules", map[string]any{"Rules": rules})
}

func (s *Server) HandleAdminModels(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	models, _ := s.Queries.ListModels(ctx)
	s.render(w, "admin_models", map[string]any{"Models": models})
}
