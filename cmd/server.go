package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"lan-notify/internal/config"
	"lan-notify/internal/discovery"
	"lan-notify/internal/i18n"
	"lan-notify/internal/notifier"
	"lan-notify/internal/security"
	"lan-notify/internal/server"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   i18n.T("cmd_server_use"),
	Short: i18n.T("cmd_server_short"),
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Configuration
		cfg, err := config.Load()
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal(i18n.T("err_config_not_found"))
			}
			log.Fatalf("%s: %v", i18n.T("err_load_config"), err)
		}

		// 2. Generate Ephemeral TLS Config
		tlsConfig, err := security.GenerateEphemeralTLSConfig(cfg.DeviceName)
		if err != nil {
			log.Fatalf("%s: %v", i18n.T("err_tls_fail"), err)
		}

		// 3. Initialize Discovery Service (mDNS)
		disc := discovery.New()
		shutdownDiscovery, err := disc.Register(cfg.DeviceName, cfg.Port)
		if err != nil {
			log.Fatalf("%s: %v", i18n.T("err_mdns_start_fail"), err)
		}
		defer shutdownDiscovery()
		log.Print(i18n.T("msg_service_registered", cfg.DeviceName, cfg.Port))

		// Print Local IPs for direct connection
		ips := getLocalIPs()
		if len(ips) > 0 {
			fmt.Println(i18n.T("msg_local_ips"))
			for _, ip := range ips {
				fmt.Print(i18n.T("msg_local_ips_arrow", ip))
			}
			fmt.Println()
		}

		// 4. Initialize Native Notifier Engine
		ntf, err := notifier.New()
		if err != nil {
			fmt.Println(i18n.T("err_fatal_init_notifier"))
			fmt.Println(i18n.T("tip_fatal_init_termux"))
			fmt.Printf(i18n.T("msg_tech_details")+": %v\n", err)
			os.Exit(1)
		}

		// 5. Initialize and Start HTTP Server (Injected dependencies)
		httpSrv := server.New(cfg, ntf, tlsConfig)

		go func() {
			if err := httpSrv.Start(); err != nil {
				log.Fatalf("%s: %v", i18n.T("err_https_server"), err)
			}
		}()

		// Wait for SIGINT/SIGTERM
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Println(i18n.T("msg_shutting_down"))
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

// getLocalIPs returns a list of non-loopback IPv4 addresses
func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}
	return ips
}
