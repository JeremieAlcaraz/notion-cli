// Package tui provides terminal UI helpers using gum (https://github.com/charmbracelet/gum).
// All functions degrade gracefully when gum is not installed.
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var (
	gumOnce      sync.Once
	gumAvailable bool
	noGum        bool // set via SetNoGum (--no-gum flag)
)

// SetNoGum forces gum to be treated as unavailable regardless of installation.
func SetNoGum(v bool) {
	noGum = v
}

// IsAvailable reports whether gum is installed and --no-gum is not set.
func IsAvailable() bool {
	if noGum {
		return false
	}
	gumOnce.Do(func() {
		_, err := exec.LookPath("gum")
		gumAvailable = err == nil
	})
	return gumAvailable
}

// WarnIfMissing prints a one-time notice to stderr if gum is not found.
// Call once at startup (e.g. in PersistentPreRun).
func WarnIfMissing() {
	if noGum {
		return
	}
	if !IsAvailable() {
		fmt.Fprintf(os.Stderr, "tip: install gum for a richer experience → brew install gum\n")
	}
}
