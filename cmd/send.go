package cmd

import (
	"context"
	"log"
	"strings"
	"time"

	"lan-notify/internal/client"
	"lan-notify/internal/config"
	"lan-notify/internal/discovery"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [message] [@target]",
	Short: "Sends a notification to a target device",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Configuration (to get the AuthToken)
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
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
			log.Fatalf("No message provided")
		}

		if target == "" {
			log.Fatalf("No target provided (use @target)")
		}

		// 3. Resolve Target via Discovery Service
		disc := discovery.New()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		resolvedIP, resolvedPort, err := disc.ResolveTarget(ctx, target)
		if err != nil {
			log.Fatalf("Target '%s' not found on local network: %v", target, err)
		}

		log.Printf("Found %s at %s:%d. Sending notification...", target, resolvedIP, resolvedPort)

		// 4. Send Notification via Client
		req := client.NotificationRequest{
			Title:   "Message from " + cfg.DeviceName,
			Message: message,
			Urgency: "normal",
		}

		if err := client.SendNotification(resolvedIP, resolvedPort, cfg.AuthToken, req); err != nil {
			log.Fatalf("Failed to send notification: %v", err)
		}

		log.Println("Notification successfully delivered!")
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
