package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"openshield-agent/internal/executor"
	"openshield-agent/internal/models"
	"openshield-agent/internal/utils"
	"strings"
)

var toolRegistry = map[string]Tool{}

// func loadToolsConfig(path string) ([]Tool, error) {
// 	data, err := os.ReadFile(path + string(os.PathSeparator) + "tools.yml")
// 	if err != nil {
// 		return nil, err
// 	}
// 	var cfg ToolsConfig
// 	if err := yaml.Unmarshal(data, &cfg); err != nil {
// 		return nil, err
// 	}
// 	return cfg.Tools, nil
// }

func RegisterTool(t Tool) {
	toolRegistry[t.Name] = t
}

// func RegisterToolsFromConfig() error {
// 	tools, err := loadToolsConfig(config.ConfigPath)
// 	if err != nil {
// 		return err
// 	}
// 	for _, t := range tools {
// 		RegisterTool(t)
// 	}
// 	return nil
// }

func GetTool(name string) (Tool, bool) {
	t, ok := toolRegistry[name]
	return t, ok
}

type Tool struct {
	Name    string   `yaml:"name"`
	Actions []Action `yaml:"actions"`
	OS      []string `yaml:"os"`
}

type Action struct {
	Name string   `yaml:"name"`
	Opts []string `yaml:"opts"`
}

type ToolsConfig struct {
	Tools []Tool `yaml:"tools"`
}

// isActionSupported checks if the given action is supported by the tool.
func (t *Tool) isActionSupported(action string) bool {
	for _, a := range t.Actions {
		if a.Name == action {
			return true
		}
	}
	return false
}

// isOSCompatible checks if the tool is compatible with the current OS.
func (t *Tool) isOSSupported(os string) bool {
	for _, supportedOS := range t.OS {
		if supportedOS == os {
			return true
		}
	}
	return false
}

// Helper to parse ID from /etc/os-release
func parseOSReleaseID(data []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			return strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
	}
	return ""
}

func (t *Tool) RunAction(action string, args []string) (string, error) {
	os := utils.GetDeviceOS()

	if !t.isActionSupported(action) {
		return "", fmt.Errorf("action %s not supported for tool %s", action, t.Name)
	}

	if !t.isOSSupported(os) {
		return "", fmt.Errorf("tool %s not supported on %s", t.Name, os)
	}

	// Execute the action
	result, err := executor.ExecuteCommand(models.Command{
		Command:  action,
		Args:     []string{action},
		TargetOS: models.OS(os),
	})
	if err != nil {
		return "", fmt.Errorf("failed to run %s: %v\n%s", action, err, result)
	}
	fmt.Printf("Tool %s executed action %s successfully\n", t.Name, action)

	return result, nil
}
