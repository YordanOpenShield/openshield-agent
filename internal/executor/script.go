package executor

import (
	"fmt"
	"regexp"
	"runtime"
)

func ExecuteScript(scriptName string) (string, error) {
	allowed := regexp.MustCompile(`^[a-zA-Z0-9_\-]+\.(sh|ps.*)$`)
	if !allowed.MatchString(scriptName) {
		return "", fmt.Errorf("invalid script name")
	}

	scriptPath := fmt.Sprintf("./scripts/%s", scriptName)
	// Detect OS and choose shell accordingly
	if runtime.GOOS == "windows" {
		return runCommand("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	}
	return runCommand("/bin/bash", scriptPath)
}
