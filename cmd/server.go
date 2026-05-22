package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"lan-notify/internal/config"
	"lan-notify/internal/notifier"
	"lan-notify/internal/security"

	"github.com/grandcat/zeroconf"
	"github.com/spf13/cobra"
)

type NotificationRequest struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Urgency  string `json:"urgency"`
	Category string `json:"category"`
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the lan-notify daemon",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("config.json")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// 1. Generate Ephemeral TLS Config
		tlsConfig, err := security.GenerateEphemeralTLSConfig(cfg.DeviceName)
		if err != nil {
			log.Fatalf("Failed to generate TLS config: %v", err)
		}

		// 2. Setup mDNS (Zeroconf)
		server, err := zeroconf.Register(cfg.DeviceName, "_lan-notifier._tcp", "local.", cfg.Port, []string{"txtv=0", "lo=1", "la=2"}, nil)
		if err != nil {
			log.Fatalf("Failed to start mDNS server: %v", err)
		}
		defer server.Shutdown()
		log.Printf("mDNS Service registered as: %s._lan-notifier._tcp.local on port %d", cfg.DeviceName, cfg.Port)

		// 3. Initialize Notifier Engine
		ntf, err := notifier.New()
		if err != nil {
			log.Fatalf("Failed to initialize native notifier: %v", err)
		}

		// 4. Setup HTTP/TLS Server
		mux := http.NewServeMux()
		mux.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			authHeader := r.Header.Get("Authorization")
			expectedAuth := "Bearer " + cfg.AuthToken
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

			if err := ntf.Notify(req.Title, req.Message, opts); err != nil {
				log.Printf("Error showing notification: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Notification sent.")
		})

		httpServer := &http.Server{
			Addr:      fmt.Sprintf(":%d", cfg.Port),
			Handler:   mux,
			TLSConfig: tlsConfig,
		}

		go func() {
			log.Printf("Listening for HTTPS requests on :%d...", cfg.Port)
			// Cert and Key paths are empty because they are loaded into TLSConfig
			if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS Server Error: %v", err)
			}
		}()

		// Wait for SIGINT/SIGTERM
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
