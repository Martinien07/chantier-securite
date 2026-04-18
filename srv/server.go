package srv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"srv.exe.dev/db/dbgen"
)

const SiteCookieName = "safesite_current_site"

type Server struct {
	DB        *sql.DB
	Queries   *dbgen.Queries
	tmpls     *template.Template
	mu        sync.Mutex
	runningAI map[int32]*exec.Cmd
}

func NewServer(database *sql.DB) (*Server, error) {
	s := &Server{
		DB:        database,
		Queries:   dbgen.New(database),
		runningAI: make(map[int32]*exec.Cmd),
	}
	if err := s.loadTemplates(); err != nil {
		return nil, err
	}
	return s, nil
}

// --- GESTION DES TEMPLATES ---

func (s *Server) loadTemplates() error {
	exePath, _ := os.Executable()
	tmplDir := filepath.Join(filepath.Dir(exePath), "templates")

	tmpls, err := template.ParseGlob(filepath.Join(tmplDir, "*.html"))
	if err != nil {
		return fmt.Errorf("parsing templates: %w", err)
	}
	s.tmpls = tmpls
	return nil
}

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

// --- ROUTAGE (Mux corrigé sans doublons) ---

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	fmt.Println(" !!! SERVEUR GO PRÊT : EN ATTENTE DE CONNEXION SUR http://localhost:8000 !!!")

	// Gestion des fichiers statiques avec MIME types corrects
	exePath, _ := os.Executable()
	staticDir := filepath.Join(filepath.Dir(exePath), "static")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".woff2", "font/woff2")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// --- PAGES PUBLIQUES / SELECTION ---
	mux.HandleFunc("GET /select-site", s.HandleSelectSite)
	mux.HandleFunc("POST /select-site", s.HandleSetSite)

	// --- PAGES PRINCIPALES (nécessitent un site) ---
	mux.HandleFunc("GET /", s.requireSite(s.HandleDashboard))
	mux.HandleFunc("GET /cameras", s.requireSite(s.HandleCameras))
	mux.HandleFunc("GET /alertes", s.requireSite(s.HandleAlerts))
	mux.HandleFunc("GET /zones", s.requireSite(s.HandleZones))
	mux.HandleFunc("GET /analyses", s.requireSite(s.HandleAnalyses))
	mux.HandleFunc("GET /rapports", s.requireSite(s.HandleRapports))
	mux.HandleFunc("GET /surveillance", s.requireSite(s.HandleSurveillance))
	mux.HandleFunc("GET /suivi", s.requireSite(s.HandleSuivi))

	mux.HandleFunc("GET /validation", s.requireSite(s.HandleValidation))

	// --- PAGES ADMIN ---
	mux.HandleFunc("GET /admin", s.requireSite(s.HandleAdmin))
	mux.HandleFunc("GET /admin/plans", s.requireSite(s.HandleAdminPlans))
	mux.HandleFunc("GET /admin/cameras", s.requireSite(s.HandleAdminCameras))
	mux.HandleFunc("GET /admin/zones", s.requireSite(s.HandleAdminZones))
	mux.HandleFunc("GET /admin/users", s.requireSite(s.HandleAdminUsers))
	mux.HandleFunc("GET /admin/rules", s.requireSite(s.HandleAdminRules))
	mux.HandleFunc("GET /admin/models", s.requireSite(s.HandleAdminModels))
	mux.HandleFunc("GET /admin/sites", s.requireAdminAuth(s.HandleAdminSites))

	// --- API DATA (Lecture) ---
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
	mux.HandleFunc("GET /api/detection/status", s.requireSiteAPI(s.HandleAPIDetectionStatus))

	// --- API VALIDATION & ALERTES (Actions) ---
	mux.HandleFunc("POST /api/alerts/{id}/acknowledge", s.HandleAPIAckAlert)
	mux.HandleFunc("POST /api/alerts/{id}/close", s.HandleAPICloseAlert)
	mux.HandleFunc("POST /api/alerts/{id}/resolve", s.HandleAPIResolveAlert)
	mux.HandleFunc("PUT /api/alerts/{id}", s.HandleAPIUpdateAlert)

	// --- API RISK EVENTS (Validation humaine) ---
	mux.HandleFunc("GET /api/pending-events", s.requireSiteAPI(s.HandleAPIPendingEvents))
	mux.HandleFunc("GET /api/risk-events/pending", s.HandleAPIPendingRiskEvents)
	mux.HandleFunc("GET /api/risk-events/pending/count", s.HandleAPIPendingCount)
	mux.HandleFunc("GET /api/pending-locations", s.requireSiteAPI(s.HandleAPIPendingLocations))
	mux.HandleFunc("POST /api/risk-events/{id}/revert-to-pending", s.HandleAPIRevertToPending)
	mux.HandleFunc("POST /api/risk-events/{id}/validate", s.HandleAPIValidateEvent)
	mux.HandleFunc("POST /api/risk-events/{id}/accept", s.HandleAPIAcceptRiskEvent)
	mux.HandleFunc("POST /api/risk-events/{id}/reject", s.HandleAPIRejectRiskEvent)

	// --- API ADMIN (CRUD) ---
	mux.HandleFunc("POST /api/admin/auth", s.HandleAPIAdminAuth)
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
	mux.HandleFunc("POST /api/admin/sites", s.requireAdminAuth(s.HandleAPICreateSite))
	mux.HandleFunc("PUT /api/admin/sites/{id}", s.requireAdminAuth(s.HandleAPIUpdateSite))
	mux.HandleFunc("DELETE /api/admin/sites/{id}", s.requireAdminAuth(s.HandleAPIDeleteSite))

	mux.HandleFunc("GET /api/cameras/status", s.requireSiteAPI(s.HandleAPICameraStatus))
	mux.HandleFunc("GET /api/ai/status", s.requireSiteAPI(s.HandleAPIAIStatus))

	mux.HandleFunc("GET /api/cameras/{id}/stream", s.HandleCameraStream)
	mux.HandleFunc("GET /api/cameras/{id}/snapshot", s.HandleCameraSnapshot)

	return mux
}

// --- MIDDLEWARES & HELPERS ---

func (s *Server) requireSite(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := s.getCurrentSiteID(r)
		if id == 0 {
			http.Redirect(w, r, "/select-site", http.StatusSeeOther)
			return
		}
		s.StartAIProcess(int32(id))
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

// --- GESTION DU PROCESSUS IA (PYTHON) ---

func (s *Server) StartAIProcess(siteID int32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.runningAI[siteID]; exists {
		return
	}

	scriptPath := os.Getenv("DETECTION_SCRIPT_PATH")
	if scriptPath == "" {
		exePath, _ := os.Executable()
		projectRoot := filepath.Dir(exePath)
		scriptPath = filepath.Join(projectRoot, "..", "TOURMANT 1", "main_ai_bridge.py")

		scriptPath = filepath.Clean(scriptPath)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		slog.Error("script Python introuvable", "path", scriptPath)
		return
	}

	pythonExe := os.Getenv("PYTHON_EXE")
	if pythonExe == "" {
		pythonExe = "py"
	}

	cmd := exec.Command(pythonExe, scriptPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("DETECTION_SITE_ID=%d", siteID))
	cmd.Dir = filepath.Dir(scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		slog.Error("impossible de démarrer l'IA", "err", err)
		return
	}

	s.runningAI[siteID] = cmd
	go func(id int32, c *exec.Cmd) {
		c.Wait()
		s.mu.Lock()
		delete(s.runningAI, id)
		s.mu.Unlock()
	}(siteID, cmd)
}

func (s *Server) StopAIProcess(siteID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cmd, exists := s.runningAI[siteID]
	if !exists {
		return fmt.Errorf("aucun processus actif pour le site %d", siteID)
	}

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("impossible d'arrêter le processus: %w", err)
	}

	delete(s.runningAI, siteID)
	return nil
}

func (s *Server) AIStatus(siteID int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.runningAI[siteID]
	return exists
}

// --- HELPERS DE CONVERSION SQL NULL ---

func ptrStr(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case *string:
		if t == nil {
			return ""
		}
		return *t
	case sql.NullString:
		if !t.Valid {
			return ""
		}
		return t.String
	default:
		return ""
	}
}

func ptrInt(v any) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return t
	case sql.NullInt32:
		if !t.Valid {
			return 0
		}
		return int64(t.Int32)
	case sql.NullInt16:
		if !t.Valid {
			return 0
		}
		return int64(t.Int16)
	case sql.NullInt64:
		if !t.Valid {
			return 0
		}
		return t.Int64
	default:
		return 0
	}
}

func ptrFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case sql.NullFloat64:
		if !t.Valid {
			return 0
		}
		return t.Float64
	default:
		return 0
	}
}

func strPtr(s string) *string     { return &s }
func intPtr(i int64) *int64       { return &i }
func floatPtr(f float64) *float64 { return &f }

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
