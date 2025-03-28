package ml

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// BaseConfig provides common configuration functionality
type BaseConfig struct {
	ConfigPath string
}

// LoadConfig loads configuration from a file, falling back to environment variables
func (c *BaseConfig) LoadConfig(configPath string, envPrefix string, config interface{}) error {
	// Try to load from file first
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := json.Unmarshal(data, config); err == nil {
				log.Printf("Loaded configuration from file: %s", configPath)
				return nil
			}
		}
	}

	// Try default config file in config directory
	defaultPath := filepath.Join("config", fmt.Sprintf("%s.json", envPrefix))
	if data, err := os.ReadFile(defaultPath); err == nil {
		if err := json.Unmarshal(data, config); err == nil {
			log.Printf("Loaded configuration from default file: %s", defaultPath)
			return nil
		}
	}

	// Fall back to environment variables
	log.Printf("Using environment variables for %s configuration", envPrefix)
	return nil
}
