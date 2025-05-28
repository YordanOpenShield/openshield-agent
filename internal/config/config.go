package config

import (
	"os"

	"runtime"

	"gopkg.in/yaml.v2"
)

var ConfigPath string = "config/config.yml"
var ScriptsPath string = "scripts"

func init() {
	switch runtime.GOOS {
	case "windows":
		ConfigPath = "C:\\ProgramData\\openshield\\config.yml"
		ScriptsPath = "C:\\ProgramData\\openshield\\scripts"
	default:
		ConfigPath = "/etc/openshield/config.yml"
		ScriptsPath = "/etc/openshield/scripts"
	}
}

var GlobalConfig Config

type Config struct {
	MANAGER_ADDRESS   string `yaml:"MANAGER_ADDRESS"`
	MANAGER_API_PORT  string `yaml:"MANAGER_API_PORT"`
	MANAGER_GRPC_PORT string `yaml:"MANAGER_GRPC_PORT"`
	COMMAND_TIMEOUT   string `yaml:"COMMAND_TIMEOUT"`
}

func GenerateConfig(managerAddress string) *Config {
	return &Config{
		MANAGER_ADDRESS:   managerAddress,
		MANAGER_API_PORT:  "9000",
		MANAGER_GRPC_PORT: "50052",
		COMMAND_TIMEOUT:   "60",
	}
}

func LoadConfig(configPath string) (*Config, error) {
	configFile := configPath + string(os.PathSeparator) + "config.yml"
	data, err := os.ReadFile(configFile)
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
