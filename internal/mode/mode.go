// Package mode holds global runtime mode flags (agent vs human).
// It has zero dependencies so any package can import it safely.
package mode

import "os"

var agent bool

// SetAgent enables or disables agent mode.
func SetAgent(v bool) { agent = v }

// IsAgent reports whether agent mode is active.
func IsAgent() bool { return agent }

// InitFromEnv checks NOTION_AGENT=1 and sets agent mode accordingly.
// Call once at startup before any command runs.
func InitFromEnv() {
	if os.Getenv("NOTION_AGENT") == "1" {
		SetAgent(true)
	}
}
