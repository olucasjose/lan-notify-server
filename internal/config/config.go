package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration.
type Config struct {
	DeviceName  string            `json:"device_name"`
	Port        int               `json:"port"`
	PinnedPeers map[string]string `json:"pinned_peers"`
}

// GetConfigDir returns the base directory for lan-notify configuration.
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "lan-notify"), nil
}

// Save writes the current configuration to the system config directory.
func (cfg *Config) Save() error {
	appDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}

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
	appDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(appDir, "config.json")
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

	if cfg.Port == 0 {
		cfg.Port = 42931
	}

	if cfg.PinnedPeers == nil {
		cfg.PinnedPeers = make(map[string]string)
	}

	return &cfg, nil
}
