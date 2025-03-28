package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all application configuration
type Config struct {
	Server struct {
		Port      string `json:"port"`
		StaticDir string `json:"static_dir"`
		Debug     bool   `json:"debug"`
	} `json:"server"`

	Database struct {
		Path string `json:"path"`
	} `json:"database"`

	ML struct {
		Type string `json:"type"` // "local" or "google"
	} `json:"ml"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Handle missing values
	if config.Server.Port == "" {
		// Fail if port is not set
		return nil, fmt.Errorf("server port is not set in config file")
	}
	if config.Server.StaticDir == "" {
		config.Server.StaticDir = "./static"
	}
	if config.Database.Path == "" {
		config.Database.Path = "nutritional.db"
	}

	return &config, nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	// First try environment variable
	if path := os.Getenv("NUTRITIONAL_CONFIG"); path != "" {
		return path
	}

	// Then try config directory
	configDir := "config"
	if _, err := os.Stat(configDir); err == nil {
		return filepath.Join(configDir, "config.json")
	}

	// Finally, try current directory
	return "config.json"
}
