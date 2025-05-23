package executor

import (
	"openshield-agent/internal/models"
	"testing"
)

func TestIsCommandWhitelisted(t *testing.T) {
	cmd := models.Command{Command: "uptime", TargetOS: models.OSLinux}
	if !IsCommandWhitelisted(cmd) {
		t.Error("Expected uptime to be whitelisted for Linux")
	}

	cmd = models.Command{Command: "notallowed", TargetOS: models.OSLinux}
	if IsCommandWhitelisted(cmd) {
		t.Error("Expected notallowed to NOT be whitelisted")
	}
}
