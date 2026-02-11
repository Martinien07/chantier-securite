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
	"strconv"
	"time"

	"srv.exe.dev/db"
	"srv.exe.dev/db/dbgen"
)

type Server struct {
	DB           *sql.DB
	Hostname     string
	TemplatesDir string
	StaticDir    string
}

func New(dbPath, hostname string) (*Server, error) {
	_, thisFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(thisFile)
	srv := &Server{
		Hostname:     hostname,
		TemplatesDir: filepath.Join(baseDir, "templates"),
		StaticDir:    filepath.Join(baseDir, "static"),
	}
	if err := srv.setUpDatabase(dbPath); err != nil {
		return nil, err
	}
	return srv, nil
}

func (s *Server) setUpDatabase(dbPath string) error {
	wdb, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	s.DB = wdb
	if err := db.RunMigrations(wdb); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Serve starts the HTTP server
func (s *Server) Serve(addr string) error {
	mux := http.NewServeMux()
	
	// Pages
	mux.HandleFunc("GET /{$}", s.HandleDashboard)
	mux.HandleFunc("GET /cameras", s.HandleCameras)
	mux.HandleFunc("GET /alertes", s.HandleAlertes)
	mux.HandleFunc("GET /analyses", s.HandleAnalyses)
	mux.HandleFunc("GET /rapports", s.HandleRapports)
	
	// API endpoints
	mux.HandleFunc("GET /api/stats", s.HandleAPIStats)
	mux.HandleFunc("GET /api/alertes", s.HandleAPIAlertes)
	mux.HandleFunc("POST /api/alertes/{id}/acknowledge", s.HandleAPIAcknowledgeAlerte)
	mux.HandleFunc("GET /api/cameras", s.HandleAPICameras)
	mux.HandleFunc("GET /api/zones", s.HandleAPIZones)
	
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.StaticDir))))
	
	slog.Info("starting server", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data any) error {
	path := filepath.Join(s.TemplatesDir, name)
	basePath := filepath.Join(s.TemplatesDir, "base.html")
	tmpl, err := template.ParseFiles(basePath, path)
	if err != nil {
		return fmt.Errorf("parse template %q: %w", name, err)
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("execute template %q: %w", name, err)
	}
	return nil
}

// Page Handlers
func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":      "Tableau de bord",
		"ActivePage": "dashboard",
	}
	if err := s.renderTemplate(w, "dashboard.html", data); err != nil {
		slog.Warn("render template", "error", err)
		http.Error(w, "Internal error", 500)
	}
}

func (s *Server) HandleCameras(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":      "Caméras",
		"ActivePage": "cameras",
	}
	if err := s.renderTemplate(w, "cameras.html", data); err != nil {
		slog.Warn("render template", "error", err)
		http.Error(w, "Internal error", 500)
	}
}

func (s *Server) HandleAlertes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":      "Alertes",
		"ActivePage": "alertes",
	}
	if err := s.renderTemplate(w, "alertes.html", data); err != nil {
		slog.Warn("render template", "error", err)
		http.Error(w, "Internal error", 500)
	}
}

func (s *Server) HandleAnalyses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":      "Analyses IA",
		"ActivePage": "analyses",
	}
	if err := s.renderTemplate(w, "analyses.html", data); err != nil {
		slog.Warn("render template", "error", err)
		http.Error(w, "Internal error", 500)
	}
}

func (s *Server) HandleRapports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":      "Rapports",
		"ActivePage": "rapports",
	}
	if err := s.renderTemplate(w, "rapports.html", data); err != nil {
		slog.Warn("render template", "error", err)
		http.Error(w, "Internal error", 500)
	}
}

// API Handlers
func (s *Server) HandleAPIStats(w http.ResponseWriter, r *http.Request) {
	q := dbgen.New(s.DB)
	
	alertesActives, _ := q.CountActiveAlertes(r.Context())
	alertesJour, _ := q.CountAlertesToday(r.Context())
	camerasActives, _ := q.CountActiveCameras(r.Context())
	zonesRisque, _ := q.CountHighRiskZones(r.Context())
	
	stats := map[string]interface{}{
		"alertes_actives": alertesActives,
		"alertes_jour":    alertesJour,
		"cameras_actives": camerasActives,
		"zones_risque":    zonesRisque,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) HandleAPIAlertes(w http.ResponseWriter, r *http.Request) {
	q := dbgen.New(s.DB)
	
	limitStr := r.URL.Query().Get("limit")
	limit := int64(50)
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	
	alertes, err := q.ListAlertes(r.Context(), limit)
	if err != nil {
		slog.Warn("list alertes", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alertes)
}

func (s *Server) HandleAPIAcknowledgeAlerte(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}
	
	now := time.Now()
	q := dbgen.New(s.DB)
	err = q.AcknowledgeAlerte(r.Context(), dbgen.AcknowledgeAlerteParams{
		ID:             id,
		AcknowledgedAt: &now,
	})
	if err != nil {
		slog.Warn("acknowledge alerte", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) HandleAPICameras(w http.ResponseWriter, r *http.Request) {
	q := dbgen.New(s.DB)
	
	cameras, err := q.ListCameras(r.Context())
	if err != nil {
		slog.Warn("list cameras", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cameras)
}

func (s *Server) HandleAPIZones(w http.ResponseWriter, r *http.Request) {
	q := dbgen.New(s.DB)
	
	zones, err := q.ListZones(r.Context())
	if err != nil {
		slog.Warn("list zones", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(zones)
}
