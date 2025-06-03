package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"openshield-agent/internal/utils"
	"strings"
)

var toolRegistry = make(map[string]Tool)

func RegisterTool(t Tool) {
	toolRegistry[t.Name] = t
}

func GetTools() map[string]Tool {
	return toolRegistry
}

func GetTool(name string) (Tool, bool) {
	if tool, exists := toolRegistry[name]; exists {
		return tool, true
	}
	return Tool{}, false
}

type Tool struct {
	Name    string   `yaml:"name" json:"name"`
	Actions []Action `yaml:"actions" json:"actions"`
	OS      []string `yaml:"os" json:"os"`
}

type Action struct {
	Name string                              `yaml:"name"`
	Opts []string                            `yaml:"opts"`
	Exec func(opts []string) (string, error) `yaml:"-"`
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

func (t *Tool) ExecAction(action string, args []string) (string, error) {
	if !t.isActionSupported(action) {
		return "", fmt.Errorf("action %s not supported by tool %s", action, t.Name)
	}

	if !t.isOSSupported(utils.GetDeviceOS()) {
		return "", fmt.Errorf("tool %s is not supported on this OS", t.Name)
	}

	for _, a := range t.Actions {
		if a.Name == action {
			if a.Exec != nil {
				return a.Exec(args)
			}
			return "", fmt.Errorf("no execution function defined for action %s in tool %s", action, t.Name)
		}
	}

	return "", fmt.Errorf("action %s not found in tool %s", action, t.Name)
}
