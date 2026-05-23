package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"time"

	"lan-notify/internal/config"
	"lan-notify/internal/security"
)

// GetSecureClient creates an http.Client configured for mTLS with TOFU/Key Pinning.
func GetSecureClient(cfg *config.Config, targetName string) (*http.Client, error) {
	appDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	clientCert, err := security.LoadOrGeneratePersistentIdentity(appDir, cfg.DeviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to load identity: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{*clientCert},
		InsecureSkipVerify: true, // We verify manually in VerifyPeerCertificate
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("no certificate presented by server")
			}

			serverCert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return err
			}

			fingerprint := security.GetCertificateFingerprint(serverCert)

			pinnedFingerprint, exists := cfg.PinnedPeers[targetName]
			if !exists {
				// We don't TOFU automatically here anymore, the pairing command does it.
				// But wait, if this is the `pair` command, it needs to accept the cert to finish pairing!
				// We can return a specific error that the pair command catches, OR we can pass a boolean `isPairing`.
				return fmt.Errorf("untrusted_server_cert:%s", fingerprint)
			}

			if fingerprint != pinnedFingerprint {
				return fmt.Errorf("security alert: server certificate fingerprint has changed (MITM attack?)")
			}

			return nil
		},
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 5 * time.Second,
	}, nil
}
