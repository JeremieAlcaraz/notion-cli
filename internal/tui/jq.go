package tui

import (
	"bytes"
	"os"
	"os/exec"
	"sync"
)

var (
	jqOnce      sync.Once
	jqAvailable bool
)

// IsJQAvailable reports whether jq is installed.
func IsJQAvailable() bool {
	jqOnce.Do(func() {
		_, err := exec.LookPath("jq")
		jqAvailable = err == nil
	})
	return jqAvailable
}

// ColorJSON prints JSON data with color via jq.
// Falls back to plain print if jq is unavailable or fails.
// Returns false if fallback was used.
func ColorJSON(data []byte) bool {
	if noGum || !IsJQAvailable() {
		return false
	}
	cmd := exec.Command("jq", "--color-output", ".")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}
