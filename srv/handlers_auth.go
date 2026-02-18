package srv

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"srv.exe.dev/db/dbgen"
)

const (
	adminUsername = "admin"
	adminPassword = "admin"
	adminCookie   = "admin_session"
)

// Middleware to require admin authentication
func (s *Server) requireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminCookie)
		if err != nil || cookie.Value != "authenticated" {
			http.Redirect(w, r, "/select-site", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// HandleAPIAdminAuth handles admin login
func (s *Server) HandleAPIAdminAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == adminUsername && req.Password == adminPassword {
		http.SetCookie(w, &http.Cookie{
			Name:     adminCookie,
			Value:    "authenticated",
			Path:     "/",
			MaxAge:   3600, // 1 hour
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// HandleAdminSites renders the sites management page
func (s *Server) HandleAdminSites(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sites, _ := s.Queries.ListSites(ctx)
	s.render(w, "admin_sites", map[string]any{"Sites": sites})
}

// HandleAPICreateSite creates a new site
func (s *Server) HandleAPICreateSite(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	site, err := s.Queries.CreateSite(ctx, dbgen.CreateSiteParams{
		Name:     req.Name,
		Location: &req.Location,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, site)
}

// HandleAPIUpdateSite updates an existing site
func (s *Server) HandleAPIUpdateSite(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	var req struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := s.Queries.UpdateSite(ctx, dbgen.UpdateSiteParams{
		ID:       id,
		Name:     req.Name,
		Location: &req.Location,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

// HandleAPIDeleteSite deletes a site
func (s *Server) HandleAPIDeleteSite(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	err := s.Queries.DeleteSite(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, map[string]string{"status": "ok"})
}
