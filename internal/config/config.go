package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

var GlobalConfig Config

type Config struct {
	MANAGER_ADDRESS   string `yaml:"MANAGER_ADDRESS"`
	MANAGER_API_PORT  string `yaml:"MANAGER_API_PORT"`
	MANAGER_GRPC_PORT string `yaml:"MANAGER_GRPC_PORT"`
	COMMAND_TIMEOUT   string `yaml:"COMMAND_TIMEOUT"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadAndSetConfig(path string) error {
	cfg, err := LoadConfig(path)
	if err != nil {
		return err
	}
	GlobalConfig = *cfg
	return nil
}
