package tools

import (
	"fmt"
	"openshield-agent/internal/executor"
	"openshield-agent/internal/utils"
)

type ScriptTool struct {
	Tool

	script string `yaml:"script"`
}

func (t *ScriptTool) RunAction(action string) (string, error) {
	if !t.isActionSupported(action) {
		return "", fmt.Errorf("action %s not supported for tool %s", action, t.Name)
	}

	os := utils.GetDeviceOS()
	if !t.isOSSupported(os) {
		return "", fmt.Errorf("tool %s not supported on %s", t.Name, os)
	}

	// Execute the script with the specified action
	result, err := executor.ExecuteScript(t.script, []string{action})
	if err != nil {
		return "", fmt.Errorf("failed to run script %s: %v\n%s", t.script, err, result)
	}

	fmt.Printf("Script %s executed action %s successfully\n", t.script, action)
	return result, nil
}
