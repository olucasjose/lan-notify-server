package config

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration.
type Config struct {
	DeviceName string `json:"device_name"`
	Port       int    `json:"port"`
	AuthToken  string `json:"auth_token"`
}

// Load reads the configuration from the given file path.
func Load(path string) (*Config, error) {
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

	return &cfg, nil
}
