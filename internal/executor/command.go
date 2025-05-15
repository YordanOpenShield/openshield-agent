package executor

import (
	"context"
	"errors"
	"log"
	"openshield-agent/internal/models"
	"os/exec"
	"time"
)

func ExecuteCommand(cmd models.Command) (string, error) {
	// Check if the command is whitelisted
	if !cmd.IsValidForCurrentOS() {
		return "", errors.New("command not valid for this OS")
	}

	// Check if the command is whitelisted
	if !IsCommandWhitelisted(cmd) {
		log.Printf("Rejected command: %+v", cmd)
		return "", errors.New("command not whitelisted")
	}

	execCmd := exec.Command(cmd.Command, cmd.Args...)
	out, err := execCmd.CombinedOutput()
	return string(out), err
}

func runCommand(command string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}
