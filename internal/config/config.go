package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration.
type Config struct {
	DeviceName string            `json:"device_name"`
	Port       int               `json:"port"`
	AuthToken  string            `json:"auth_token"`
	KnownPeers map[string]string `json:"known_peers"`
}

// Save writes the current configuration to the system config directory.
func (cfg *Config) Save() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(configDir, "lan-notify")
	path := filepath.Join(appDir, "config.json")

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

// Load reads the configuration from the system config directory (~/.config/lan-notify/config.json).
func Load() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(configDir, "lan-notify", "config.json")

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if cfg.KnownPeers == nil {
		cfg.KnownPeers = make(map[string]string)
	}

	return &cfg, nil
}
