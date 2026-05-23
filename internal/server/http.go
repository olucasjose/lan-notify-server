package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"lan-notify/internal/config"
	"lan-notify/internal/i18n"
	"lan-notify/internal/notifier"
	"lan-notify/internal/security"
)

// NotificationRequest defines the structure of the incoming JSON payload.
type NotificationRequest struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Urgency  string `json:"urgency"`
	Category string `json:"category"`
}

// PairRequest represents the payload for the pairing endpoint.
type PairRequest struct {
	PIN        string `json:"pin"`
	DeviceName string `json:"device_name"`
}

// HTTPServer handles incoming notification requests.
type HTTPServer struct {
	cfg       *config.Config
	ntf       notifier.Notifier
	tlsConfig *tls.Config
}

// New creates a new instance of HTTPServer.
func New(cfg *config.Config, ntf notifier.Notifier, tlsConfig *tls.Config) *HTTPServer {
	return &HTTPServer{
		cfg:       cfg,
		ntf:       ntf,
		tlsConfig: tlsConfig,
	}
}

// Start begins listening for HTTPS requests. This method blocks.
func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/notify", s.handleNotify)
	mux.HandleFunc("/pair", s.handlePair)

	httpServer := &http.Server{
		Addr:      fmt.Sprintf(":%d", s.cfg.Port),
		Handler:   mux,
		TLSConfig: s.tlsConfig,
	}

	log.Print(i18n.T("msg_listening_https", s.cfg.Port))
	return httpServer.ListenAndServeTLS("", "")
}

func (s *HTTPServer) getClientFingerprint(r *http.Request) string {
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		return security.GetCertificateFingerprint(r.TLS.PeerCertificates[0])
	}
	return ""
}

func (s *HTTPServer) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	fingerprint := s.getClientFingerprint(r)
	if fingerprint == "" {
		http.Error(w, "Unauthorized: Client certificate required", http.StatusUnauthorized)
		return
	}

	isPinned := false
	for _, pinnedFp := range s.cfg.PinnedPeers {
		if pinnedFp == fingerprint {
			isPinned = true
			break
		}
	}

	if !isPinned {
		http.Error(w, "Unauthorized: Device not paired", http.StatusUnauthorized)
		return
	}

	var req NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	opts := notifier.NotifyOptions{
		Urgency:  req.Urgency,
		Category: req.Category,
	}

	if err := s.ntf.Notify(req.Title, req.Message, opts); err != nil {
		log.Printf("Error showing notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notification sent.")
}

type PairingPinData struct {
	PIN       string    `json:"pin"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *HTTPServer) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	fingerprint := s.getClientFingerprint(r)
	if fingerprint == "" {
		http.Error(w, "Unauthorized: Client certificate required", http.StatusUnauthorized)
		return
	}

	var req PairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	appDir, err := config.GetConfigDir()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pinPath := filepath.Join(appDir, "pairing_pin.json")
	pinDataBytes, err := os.ReadFile(pinPath)
	if err != nil {
		http.Error(w, "No active pairing session", http.StatusForbidden)
		return
	}

	var pinData PairingPinData
	if err := json.Unmarshal(pinDataBytes, &pinData); err != nil {
		os.Remove(pinPath)
		http.Error(w, "Invalid pairing state", http.StatusInternalServerError)
		return
	}

	if time.Now().After(pinData.ExpiresAt) {
		os.Remove(pinPath)
		http.Error(w, "Pairing PIN expired", http.StatusForbidden)
		return
	}

	if req.PIN != pinData.PIN {
		http.Error(w, "Invalid PIN", http.StatusForbidden)
		return
	}

	// PIN is correct! Delete the file and save the peer
	os.Remove(pinPath)

	s.cfg.PinnedPeers[req.DeviceName] = fingerprint
	if err := s.cfg.Save(); err != nil {
		log.Printf("Failed to save config after pairing: %v", err)
		http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully paired with device: %s", req.DeviceName)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Pairing successful")
}
