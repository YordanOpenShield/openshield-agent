package utils

import (
	"fmt"
	"os/exec"
)

func StartSystemdService() error {
	if err := exec.Command("systemctl", "enable", "openshield-agent").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}
	if err := exec.Command("systemctl", "start", "openshield-agent").Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}
