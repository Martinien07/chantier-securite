// srv/handlers_stream.go
package srv

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// getFrameDir retourne le dossier data/frames du projet Python.
// Utilise DETECTION_SCRIPT_PATH pour déduire le chemin automatiquement.
func getFrameDir() string {
	// hse_acquisition_manager.py est dans inference/
	// Il calcule frame_dir = os.path.dirname(__file__) / data / frames
	// = TOURMANT 1/inference/data/frames/
	scriptPath := os.Getenv("DETECTION_SCRIPT_PATH")
	if scriptPath != "" {
		// DETECTION_SCRIPT_PATH = .../TOURMANT 1/system_automatic/main_ai_bridge.py
		// inference/ est au même niveau que system_automatic/
		tourmant1 := filepath.Dir(filepath.Dir(scriptPath))
		return filepath.Join(tourmant1, "inference", "data", "frames")
	}
	// Fallback : chemin absolu connu
	return `D:\ETUDEµ\CITE\ETAPE4\PROJET DE FIN DE SESSION\TOURMANT 1\inference\data\frames`
}

// HandleCameraStream — GET /api/cameras/{id}/stream
// Flux MJPEG : Python écrit cam_{id}.jpg, Go le relit en boucle.
func (s *Server) HandleCameraStream(w http.ResponseWriter, r *http.Request) {
	camID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id invalide", 400)
		return
	}

	frameDir := getFrameDir()
	framePath := filepath.Join(frameDir, fmt.Sprintf("cam_%d.jpg", camID))

	boundary := "safesite_frame"
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+boundary)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ticker := time.NewTicker(66 * time.Millisecond) // ~15 FPS
	defer ticker.Stop()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			data, err := os.ReadFile(framePath)
			if err != nil {
				data = placeholderJPEG()
			}
			fmt.Fprintf(w, "--%s\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n",
				boundary, len(data))
			w.Write(data)
			fmt.Fprintf(w, "\r\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// HandleCameraSnapshot — GET /api/cameras/{id}/snapshot
// Dernière frame JPEG statique pour l'aperçu du dashboard.
func (s *Server) HandleCameraSnapshot(w http.ResponseWriter, r *http.Request) {
	camID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id invalide", 400)
		return
	}

	frameDir := getFrameDir()
	framePath := filepath.Join(frameDir, fmt.Sprintf("cam_%d.jpg", camID))

	data, err := os.ReadFile(framePath)
	if err != nil {
		http.Error(w, "frame non disponible", 404)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write(data)
}

// placeholderJPEG : 1x1 pixel gris quand la frame n'est pas disponible.
func placeholderJPEG() []byte {
	return []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
		0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
		0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
		0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
		0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
		0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00,
		0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0xFF, 0xC4, 0x00, 0xB5, 0x10, 0x00, 0x02, 0x01, 0x03,
		0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7D,
		0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00, 0xFB, 0xFF,
		0xD9,
	}
}

// ── Routes à ajouter dans server.go ──────────────────────────────────────────
// mux.HandleFunc("GET /api/cameras/{id}/stream",   s.HandleCameraStream)
// mux.HandleFunc("GET /api/cameras/{id}/snapshot", s.HandleCameraSnapshot)
