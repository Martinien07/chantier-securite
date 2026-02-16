package srv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"srv.exe.dev/db/dbgen"
)

const SiteCookieName = "safesite_current_site"

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

	// Site selection
	mux.HandleFunc("GET /select-site", s.HandleSelectSite)
	mux.HandleFunc("POST /select-site", s.HandleSetSite)

	// Pages (require site)
	mux.HandleFunc("GET /", s.requireSite(s.HandleDashboard))
	mux.HandleFunc("GET /cameras", s.requireSite(s.HandleCameras))
	mux.HandleFunc("GET /alertes", s.requireSite(s.HandleAlerts))
	mux.HandleFunc("GET /zones", s.requireSite(s.HandleZones))
	mux.HandleFunc("GET /analyses", s.requireSite(s.HandleAnalyses))
	mux.HandleFunc("GET /rapports", s.requireSite(s.HandleRapports))

	// Admin pages
	mux.HandleFunc("GET /admin", s.requireSite(s.HandleAdmin))
	mux.HandleFunc("GET /admin/plans", s.requireSite(s.HandleAdminPlans))
	mux.HandleFunc("GET /admin/cameras", s.requireSite(s.HandleAdminCameras))
	mux.HandleFunc("GET /admin/zones", s.requireSite(s.HandleAdminZones))
	mux.HandleFunc("GET /admin/users", s.requireSite(s.HandleAdminUsers))
	mux.HandleFunc("GET /admin/rules", s.requireSite(s.HandleAdminRules))
	mux.HandleFunc("GET /admin/models", s.requireSite(s.HandleAdminModels))

	// API endpoints
	mux.HandleFunc("GET /api/stats", s.requireSiteAPI(s.HandleAPIStats))
	mux.HandleFunc("GET /api/alerts", s.requireSiteAPI(s.HandleAPIAlerts))
	mux.HandleFunc("GET /api/alerts/active", s.requireSiteAPI(s.HandleAPIActiveAlerts))
	mux.HandleFunc("GET /api/cameras", s.requireSiteAPI(s.HandleAPICameras))
	mux.HandleFunc("GET /api/sites", s.HandleAPISites)
	mux.HandleFunc("GET /api/sites/search", s.HandleAPISearchSites)
	mux.HandleFunc("GET /api/zones", s.requireSiteAPI(s.HandleAPIZones))
	mux.HandleFunc("GET /api/plans", s.requireSiteAPI(s.HandleAPIPlans))
	mux.HandleFunc("GET /api/plans/{id}/data", s.HandleAPIPlanData)
	mux.HandleFunc("GET /api/detections/recent", s.requireSiteAPI(s.HandleAPIRecentDetections))
	mux.HandleFunc("POST /api/alerts/{id}/acknowledge", s.HandleAPIAckAlert)
	mux.HandleFunc("POST /api/alerts/{id}/close", s.HandleAPICloseAlert)

	// Admin API
	mux.HandleFunc("POST /api/admin/plans", s.HandleAPICreatePlan)
	mux.HandleFunc("POST /api/admin/plans/{id}/upload", s.HandleAPIUploadPlanImage)
	mux.HandleFunc("DELETE /api/admin/plans/{id}", s.HandleAPIDeletePlan)
	mux.HandleFunc("POST /api/admin/cameras", s.HandleAPICreateCamera)
	mux.HandleFunc("PUT /api/admin/cameras/{id}", s.HandleAPIUpdateCamera)
	mux.HandleFunc("DELETE /api/admin/cameras/{id}", s.HandleAPIDeleteCamera)
	mux.HandleFunc("POST /api/admin/zones", s.HandleAPICreateZone)
	mux.HandleFunc("GET /api/admin/zones/{id}", s.HandleAPIGetZone)
	mux.HandleFunc("PUT /api/admin/zones/{id}", s.HandleAPIUpdateZone)
	mux.HandleFunc("DELETE /api/admin/zones/{id}", s.HandleAPIDeleteZone)

	return mux
}

// Middleware to require site selection
func (s *Server) requireSite(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.getCurrentSiteID(r) == 0 {
			http.Redirect(w, r, "/select-site", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (s *Server) requireSiteAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.getCurrentSiteID(r) == 0 {
			http.Error(w, "No site selected", http.StatusBadRequest)
			return
		}
		next(w, r)
	}
}

func (s *Server) getCurrentSiteID(r *http.Request) int64 {
	cookie, err := r.Cookie(SiteCookieName)
	if err != nil {
		return 0
	}
	id, _ := strconv.ParseInt(cookie.Value, 10, 64)
	return id
}

func (s *Server) getCurrentSite(r *http.Request) *dbgen.Site {
	id := s.getCurrentSiteID(r)
	if id == 0 {
		return nil
	}
	site, err := s.Queries.GetSite(r.Context(), id)
	if err != nil {
		return nil
	}
	return &site
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

func strPtr(s string) *string {
	return &s
}

func intPtr(i int64) *int64 {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

// Save uploaded file
func saveUploadedFile(r *http.Request, fieldName, destDir string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return "", fmt.Errorf("invalid file type: %s", ext)
	}

	filename := fmt.Sprintf("%d%s", header.Size, ext)
	destPath := filepath.Join(destDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filename, nil
}
