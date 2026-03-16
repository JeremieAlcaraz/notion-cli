package tui

import (
	"os"
	"os/exec"
)

// Spinner starts a gum spinner on stderr and returns a stop function.
// If gum is unavailable or stdout is not a TTY, stop is a no-op.
//
// Usage:
//
//	stop := tui.StartSpinner("Loading...")
//	defer stop()
func StartSpinner(title string) func() {
	if !IsAvailable() || !isTTY() {
		return func() {}
	}

	cmd := exec.Command("gum", "spin",
		"--spinner", "dot",
		"--title", " "+title,
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return func() {}
	}

	return func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
}

// isTTY reports whether stderr is a terminal (spinner should only appear there).
func isTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
