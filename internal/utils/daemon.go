package utils

import (
	"fmt"
	"os"
	"os/exec"
)

const systemdUnit = `[Unit]
Description=OpenShield Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/openshield-agent -manager <manager-address>
Restart=always
User=root

[Install]
WantedBy=multi-user.target
`

func InstallSystemdService() error {
	unitPath := "/etc/systemd/system/openshield-agent.service"
	if err := os.WriteFile(unitPath, []byte(systemdUnit), 0644); err != nil {
		return fmt.Errorf("failed to write unit file: %w", err)
	}
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}
	if err := exec.Command("systemctl", "enable", "openshield-agent").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}
	if err := exec.Command("systemctl", "start", "openshield-agent").Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}
