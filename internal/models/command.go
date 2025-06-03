package models

type Command struct {
	Command  string   `json:"command"`        // e.g. "uptime"
	Args     []string `json:"args,omitempty"` // e.g. ["-h"]
	TargetOS string   `json:"os"`             // e.g. "linux" or "windows"
}
