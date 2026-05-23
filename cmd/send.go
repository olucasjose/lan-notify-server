package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"lan-notify/internal/client"
	"lan-notify/internal/config"
	"lan-notify/internal/discovery"
	"lan-notify/internal/i18n"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   i18n.T("cmd_send_use"),
	Short: i18n.T("cmd_send_short"),
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Configuration
		cfg, err := config.Load()
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal(i18n.T("err_config_not_found"))
			}
			log.Fatalf("%s: %v", i18n.T("err_load_config"), err)
		}

		var target string
		var messageParts []string

		// 2. Parse Arguments
		for _, arg := range args {
			if strings.HasPrefix(arg, "@") {
				target = strings.TrimPrefix(arg, "@")
			} else {
				messageParts = append(messageParts, arg)
			}
		}

		message := strings.Join(messageParts, " ")
		if message == "" {
			log.Fatal(i18n.T("err_no_message"))
		}

		if target == "" {
			log.Fatal(i18n.T("err_no_target"))
		}

		var resolvedIP string
		var resolvedPort int

		// 3. Resolve Target (Direct IP or mDNS)
		if net.ParseIP(target) != nil {
			resolvedIP = target
			resolvedPort = cfg.Port
		} else {
			disc := discovery.New()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			var err error
			resolvedIP, resolvedPort, err = disc.ResolveTarget(ctx, target)
			if err != nil {
				fmt.Print(i18n.T("err_target_not_found", target))
				fmt.Println(i18n.T("err_target_not_found_tips"))
				os.Exit(1)
			}
		}

		req := client.NotificationRequest{
			Title:   "Message from " + cfg.DeviceName,
			Message: message,
			Urgency: "normal",
		}

		// 4. Try Sending
		err = client.SendNotification(cfg, resolvedIP, resolvedPort, target, req)

		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				fmt.Print(i18n.T("err_conn_refused", resolvedIP, resolvedPort))
				fmt.Println(i18n.T("tip_conn_refused"))
			} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Timeout") {
				fmt.Print(i18n.T("err_conn_timeout", resolvedIP, resolvedPort))
				fmt.Println(i18n.T("tip_conn_timeout"))
			} else {
				fmt.Printf("%s: %v\n", i18n.T("err_send_fail"), err)
			}
			os.Exit(1)
		}

		log.Println(i18n.T("success_send"))
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
