package models

import (
	"log"
	"runtime"
)

type OS string

const (
	OSLinux   OS = "linux"
	OSWindows OS = "windows"
)

type Command struct {
	Command  string   `json:"command"`        // e.g. "uptime"
	Args     []string `json:"args,omitempty"` // e.g. ["-h"]
	TargetOS OS       `json:"os"`             // e.g. "linux" or "windows"
}

// IsValidForCurrentOS checks if the command is valid for the current OS.
func (c *Command) IsValidForCurrentOS() bool {
	current := GetCurrentOS()
	log.Printf("[AGENT] Current OS: %s, Command Target OS: %s", current, c.TargetOS)
	return c.TargetOS == current
}

// GetCurrentOS returns the current operating system.
func GetCurrentOS() OS {
	switch os := runtime.GOOS; os {
	case "windows":
		return OSWindows
	case "linux":
		return OSLinux
	default:
		return OS("") // Unknown / unsupported
	}
}
