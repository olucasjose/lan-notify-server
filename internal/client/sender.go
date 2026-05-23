package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"lan-notify/internal/config"
)

// NotificationRequest defines the structure of the outgoing JSON payload.
type NotificationRequest struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Urgency  string `json:"urgency"`
	Category string `json:"category"`
}

// SendNotification dispatches a notification securely to the target IP and Port using mTLS.
func SendNotification(cfg *config.Config, ip string, port int, targetName string, req NotificationRequest) error {
	client, err := GetSecureClient(cfg, targetName)
	if err != nil {
		return err
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://%s:%d/notify", ip, port)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: device not paired. run 'lan-notify pair connect @%s' first", targetName)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server rejected the request with status: %s", resp.Status)
	}

	return nil
}
