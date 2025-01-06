package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application's configuration
type Config struct {
	KNIRVPath string `json:"knirv_path"`
	DBPath    string `json:"dbpath"`
	Port      string `json:"port"`
	Installed bool   `json:"installed"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Installed: false}, nil // return default config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return config, nil
}

// SaveConfig saves the configuration to a file
func (c *Config) SaveConfig(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config to file: %w", err)
	}
	return nil
}
