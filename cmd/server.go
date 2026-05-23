package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"lan-notify/internal/config"
	"lan-notify/internal/discovery"
	"lan-notify/internal/notifier"
	"lan-notify/internal/security"
	"lan-notify/internal/server"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the lan-notify daemon",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Configuration
		cfg, err := config.Load()
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("❌ Configuração não encontrada em ~/.config/lan-notify.\n💡 Dica: Rode o comando 'lan-notify config' para configurar seu dispositivo pela primeira vez.")
			}
			log.Fatalf("Failed to load config: %v", err)
		}

		// 2. Generate Ephemeral TLS Config
		tlsConfig, err := security.GenerateEphemeralTLSConfig(cfg.DeviceName)
		if err != nil {
			log.Fatalf("Failed to generate TLS config: %v", err)
		}

		// 3. Initialize Discovery Service (mDNS)
		disc := discovery.New()
		shutdownDiscovery, err := disc.Register(cfg.DeviceName, cfg.Port)
		if err != nil {
			log.Fatalf("Failed to start mDNS server: %v", err)
		}
		defer shutdownDiscovery()
		log.Printf("Service registered via mDNS as: %s on port %d", cfg.DeviceName, cfg.Port)

		// 4. Initialize Native Notifier Engine
		ntf, err := notifier.New()
		if err != nil {
			log.Fatalf("Failed to initialize native notifier: %v", err)
		}

		// 5. Initialize and Start HTTP Server (Injected dependencies)
		httpSrv := server.New(cfg, ntf, tlsConfig)

		go func() {
			if err := httpSrv.Start(); err != nil {
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
