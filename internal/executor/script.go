package executor

import (
	"fmt"
	"regexp"
)

func runScript(scriptName string) (string, error) {
	allowed := regexp.MustCompile(`^[a-zA-Z0-9_\-]+\.(sh|ps.*)$`)
	if !allowed.MatchString(scriptName) {
		return "", fmt.Errorf("invalid script name")
	}

	scriptPath := fmt.Sprintf("./scripts/%s", scriptName)
	return runCommand("/bin/bash", scriptPath)
}
