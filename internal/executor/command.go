package executor

import (
	"context"
	"errors"
	"log"
	"openshield-agent/internal/config"
	"openshield-agent/internal/models"
	"os/exec"
	"strconv"
	"time"
)

// ExecuteCommand executes a command with the given arguments and returns the output.
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

	out, err := runCommand(cmd.Command, cmd.Args...)
	return out, err
}

// runCommand executes a command with a timeout.
func runCommand(command string, args ...string) (string, error) {
	config := config.GlobalConfig

	timeoutStr := config.COMMAND_TIMEOUT
	timeout := 30 // default timeout in seconds
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}
