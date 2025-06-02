package utils

import (
	"log"
	"openshield-agent/internal/config"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func CreateConfig(configDir string, managerAddress string) {
	configFile := filepath.Join(configDir, "config.yml")

	// Check if the config directory exists, if not, create it
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

	// If the config file exists, update the manager address if needed
	if _, err := os.Stat(configFile); err == nil && managerAddress != "" {
		fileBytes, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}

		var cfg config.Config
		if err := yaml.Unmarshal(fileBytes, &cfg); err != nil {
			log.Fatalf("Failed to unmarshal config file: %v", err)
		}

		if cfg.MANAGER_ADDRESS != managerAddress {
			cfg.MANAGER_ADDRESS = managerAddress
			yamlBytes, err := yaml.Marshal(cfg)
			if err != nil {
				log.Fatalf("Failed to marshal updated config: %v", err)
			}
			err = os.WriteFile(configFile, yamlBytes, 0644)
			if err != nil {
				log.Fatalf("Failed to update config.yml: %v", err)
			}
			log.Printf("Updated manager address in config at %s", configFile)
		}
	}
}

func CreateScriptsDir(scriptsDir string) {
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		err := os.MkdirAll(scriptsDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create scripts directory: %v", err)
		}
		log.Printf("Created scripts directory at %s", scriptsDir)
	}
}

func CreateCertsDir(certsDir string) {
	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		err := os.MkdirAll(certsDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create certs directory: %v", err)
		}
		log.Printf("Created scripts directory at %s", certsDir)
	}
}
