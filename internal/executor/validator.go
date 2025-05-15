package executor

import (
	"openshield-agent/internal/models"
)

// CommandWhitelist is a list of whitelisted commands that can be executed by the agent.
var CommandWhitelist = []models.Command{
	{Command: "uptime", TargetOS: models.OSLinux},
	{Command: "df", TargetOS: models.OSLinux},
	{Command: "tasklist", TargetOS: models.OSWindows},
	{Command: "whoami", TargetOS: models.OSLinux},
	{Command: "timeout", TargetOS: models.OSWindows},
	{Command: "ping", TargetOS: models.OSWindows},
}

// IsCommandWhitelisted checks if the command is in the whitelist.
func IsCommandWhitelisted(cmd models.Command) bool {
	for _, allowed := range CommandWhitelist {
		if cmd.Command == allowed.Command &&
			cmd.TargetOS == allowed.TargetOS {
			return true
		}
	}
	return false
}
