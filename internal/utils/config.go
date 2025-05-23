package utils

import (
	"log"
	"openshield-agent/internal/config"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func CreateConfig(configFile string, managerAddress string) {
	// Check if the config directory exists, if not, create it
	configDir := filepath.Dir(configFile)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}
	}

	// Check if the config file exists, if not, create it with default values
	if _, err := os.Stat(configFile); os.IsNotExist(err) {

		if managerAddress == "" {
			log.Fatalf("manager address cannot be empty")
		}

		defaultConfig := config.GenerateConfig(managerAddress)
		// Marshal the config to YAML (or JSON if preferred)
		yamlBytes, err := yaml.Marshal(defaultConfig)
		if err != nil {
			log.Fatalf("Failed to marshal default config: %v", err)
		}
		err = os.WriteFile(configFile, yamlBytes, 0644)
		if err != nil {
			log.Fatalf("Failed to create default config.yml: %v", err)
		}
		log.Printf("Created default config at %s", configFile)
	}
}
