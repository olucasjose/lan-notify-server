package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"lan-notify/internal/config"
	"lan-notify/internal/notifier"
)

// NotificationRequest defines the structure of the incoming JSON payload.
type NotificationRequest struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Urgency  string `json:"urgency"`
	Category string `json:"category"`
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

	httpServer := &http.Server{
		Addr:      fmt.Sprintf(":%d", s.cfg.Port),
		Handler:   mux,
		TLSConfig: s.tlsConfig,
	}

	log.Printf("Listening for HTTPS requests on :%d...", s.cfg.Port)
	return httpServer.ListenAndServeTLS("", "")
}

func (s *HTTPServer) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	expectedAuth := "Bearer " + s.cfg.AuthToken
	if authHeader != expectedAuth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
