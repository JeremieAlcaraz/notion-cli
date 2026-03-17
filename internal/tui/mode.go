package tui

import "github.com/JeremieAlcaraz/notion-cli/internal/mode"

// SetAgentMode enables or disables agent mode and implies --no-gum.
func SetAgentMode(v bool) {
	mode.SetAgent(v)
	if v {
		noGum = true
	}
}

// IsAgentMode reports whether agent mode is active.
func IsAgentMode() bool { return mode.IsAgent() }

// InitAgentMode checks NOTION_AGENT env var and sets agent mode accordingly.
func InitAgentMode() { mode.InitFromEnv() }
