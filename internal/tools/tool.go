package tools

import (
	"fmt"
	"openshield-agent/internal/executor"
	"openshield-agent/internal/models"
	"openshield-agent/internal/utils"
)

type Tool struct {
	Name    string   `yaml:"name"`
	Script  string   `yaml:"script"`
	Actions []Action `yaml:"actions"`
	OS      []string `yaml:"os"`
}

type Action struct {
	Name string   `yaml:"name"`
	Opts []string `yaml:"opts"`
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
