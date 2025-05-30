package executor

import (
	"fmt"
	"openshield-agent/internal/config"
	"regexp"
	"runtime"
)

func ExecuteScript(scriptName string, args []string) (string, error) {
	allowed := regexp.MustCompile(`^[a-zA-Z0-9_\-]+\.(sh|ps.*)$`)
	if !allowed.MatchString(scriptName) {
		return "", fmt.Errorf("invalid script name")
	}

	scriptPath := fmt.Sprintf("%s/%s", config.ScriptsPath, scriptName)
	// Detect OS and choose shell accordingly
	if runtime.GOOS == "windows" {
		psArgs := append([]string{"-ExecutionPolicy", "Bypass", "-File", scriptPath}, args...)
		return runCommand("powershell", psArgs...)
	}
	return runCommand("/bin/bash", append([]string{scriptPath}, args...)...)
}
