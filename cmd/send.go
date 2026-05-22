package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"lan-notify/internal/config"

	"github.com/grandcat/zeroconf"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [message] [@target]",
	Short: "Sends a notification to a target device",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("config.json")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		var target string
		var messageParts []string

		// Parse arguments
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

		// Resolve target IP via mDNS
		resolver, err := zeroconf.NewResolver(nil)
		if err != nil {
			log.Fatalf("Failed to initialize mDNS resolver: %v", err)
		}

		entries := make(chan *zeroconf.ServiceEntry)
		var resolvedIP string
		var resolvedPort int

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		go func(results <-chan *zeroconf.ServiceEntry) {
			for entry := range results {
				if entry.Instance == target {
					if len(entry.AddrIPv4) > 0 {
						resolvedIP = entry.AddrIPv4[0].String()
						resolvedPort = entry.Port
						cancel() // Stop searching once found
					}
				}
			}
		}(entries)

		err = resolver.Browse(ctx, "_lan-notifier._tcp", "local.", entries)
		if err != nil {
			log.Fatalf("Failed to browse mDNS: %v", err)
		}

		<-ctx.Done()

		if resolvedIP == "" {
			log.Fatalf("Target '%s' not found on local network", target)
		}

		log.Printf("Found %s at %s:%d. Sending notification...", target, resolvedIP, resolvedPort)

		// Create HTTPS client with InsecureSkipVerify
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 5 * time.Second,
		}

		reqBody, _ := json.Marshal(NotificationRequest{
			Title:   "Message from " + cfg.DeviceName,
			Message: message,
			Urgency: "normal",
		})

		url := fmt.Sprintf("https://%s:%d/notify", resolvedIP, resolvedPort)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to send notification: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Server rejected the request. Status: %s", resp.Status)
		}

		log.Println("Notification successfully delivered!")
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
