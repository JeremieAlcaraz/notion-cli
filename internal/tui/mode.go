package tui

import "os"

var agentMode bool

// SetAgentMode enables or disables agent mode globally.
func SetAgentMode(v bool) {
	agentMode = v
	if v {
		// Agent mode implies --no-gum
		noGum = true
	}
}

// IsAgentMode reports whether agent mode is active.
func IsAgentMode() bool {
	return agentMode
}

// InitAgentMode checks NOTION_AGENT env var and sets agent mode accordingly.
// Call once at startup before any command runs.
func InitAgentMode() {
	if os.Getenv("NOTION_AGENT") == "1" {
		SetAgentMode(true)
	}
}
