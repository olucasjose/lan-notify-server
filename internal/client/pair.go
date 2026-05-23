package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"lan-notify/internal/config"
	"lan-notify/internal/security"
)

type PairRequest struct {
	PIN        string `json:"pin"`
	DeviceName string `json:"device_name"`
}

// ConnectPair sends the PIN to the target server to complete mTLS pairing.
func ConnectPair(cfg *config.Config, ip string, port int, targetName string, pin string) (string, error) {
	client, err := GetSecureClient(cfg, targetName)
	if err != nil {
		return "", err
	}

	reqPayload := PairRequest{
		PIN:        pin,
		DeviceName: cfg.DeviceName,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s:%d/pair", ip, port)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	
	serverFingerprint := ""

	if err != nil {
		// If we got untrusted_server_cert, it means the connection succeeded but verification failed.
		// Since we are pairing, we WANT to extract this fingerprint and proceed.
		if strings.Contains(err.Error(), "untrusted_server_cert:") {
			parts := strings.Split(err.Error(), "untrusted_server_cert:")
			if len(parts) == 2 {
				serverFingerprint = strings.TrimSpace(parts[1])
				
				// We need to temporarily trust this fingerprint to finish the pairing HTTP request.
				// So we add it to the map and retry.
				cfg.PinnedPeers[targetName] = serverFingerprint
				return ConnectPair(cfg, ip, port, targetName, pin)
			}
		}
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server rejected the pairing with status: %s", resp.Status)
	}

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		serverFingerprint = security.GetCertificateFingerprint(resp.TLS.PeerCertificates[0])
	} else if fp, ok := cfg.PinnedPeers[targetName]; ok {
		serverFingerprint = fp
	}

	return serverFingerprint, nil
}
