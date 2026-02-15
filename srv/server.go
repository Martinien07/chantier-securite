package srv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"runtime"

	"srv.exe.dev/db/dbgen"
)

type Server struct {
	DB      *sql.DB
	Queries *dbgen.Queries
	tmpls   *template.Template
}

func NewServer(database *sql.DB) (*Server, error) {
	s := &Server{
		DB:      database,
		Queries: dbgen.New(database),
	}
	if err := s.loadTemplates(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) loadTemplates() error {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	tmplDir := filepath.Join(dir, "templates")
	tmpls, err := template.ParseGlob(filepath.Join(tmplDir, "*.html"))
	if err != nil {
		return fmt.Errorf("parsing templates: %w", err)
	}
	s.tmpls = tmpls
	return nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	
	// Static files
	_, file, _, _ := runtime.Caller(0)
	staticDir := filepath.Join(filepath.Dir(file), "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	
	// Pages
	mux.HandleFunc("GET /", s.HandleDashboard)
	mux.HandleFunc("GET /cameras", s.HandleCameras)
	mux.HandleFunc("GET /alertes", s.HandleAlerts)
	mux.HandleFunc("GET /zones", s.HandleZones)
	mux.HandleFunc("GET /analyses", s.HandleAnalyses)
	mux.HandleFunc("GET /rapports", s.HandleRapports)
	
	// Admin pages
	mux.HandleFunc("GET /admin", s.HandleAdmin)
	mux.HandleFunc("GET /admin/sites", s.HandleAdminSites)
	mux.HandleFunc("GET /admin/cameras", s.HandleAdminCameras)
	mux.HandleFunc("GET /admin/users", s.HandleAdminUsers)
	mux.HandleFunc("GET /admin/rules", s.HandleAdminRules)
	mux.HandleFunc("GET /admin/models", s.HandleAdminModels)
	
	// API endpoints
	mux.HandleFunc("GET /api/stats", s.HandleAPIStats)
	mux.HandleFunc("GET /api/alerts", s.HandleAPIAlerts)
	mux.HandleFunc("GET /api/cameras", s.HandleAPICameras)
	mux.HandleFunc("GET /api/sites", s.HandleAPISites)
	mux.HandleFunc("GET /api/zones", s.HandleAPIZones)
	mux.HandleFunc("GET /api/plans/{siteId}", s.HandleAPIPlans)
	mux.HandleFunc("GET /api/detections/recent", s.HandleAPIRecentDetections)
	mux.HandleFunc("POST /api/alerts/{id}/acknowledge", s.HandleAPIAckAlert)
	mux.HandleFunc("POST /api/alerts/{id}/close", s.HandleAPICloseAlert)
	
	// Admin API
	mux.HandleFunc("POST /api/admin/cameras", s.HandleAPICreateCamera)
	mux.HandleFunc("PUT /api/admin/cameras/{id}", s.HandleAPIUpdateCamera)
	mux.HandleFunc("DELETE /api/admin/cameras/{id}", s.HandleAPIDeleteCamera)
	mux.HandleFunc("POST /api/admin/sites", s.HandleAPICreateSite)
	mux.HandleFunc("POST /api/admin/zones", s.HandleAPICreateZone)
	
	return mux
}

// Helper to render templates
func (s *Server) render(w http.ResponseWriter, tmpl string, data any) {
	if err := s.tmpls.ExecuteTemplate(w, tmpl, data); err != nil {
		slog.Error("template error", "err", err)
		http.Error(w, "Internal error", 500)
	}
}

func (s *Server) jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrInt(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func ptrFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
