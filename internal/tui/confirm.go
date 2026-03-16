package tui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/4ier/notion-cli/internal/mode"
)

// Confirm asks the user to confirm a destructive action using gum confirm.
// Returns true if confirmed, false if declined or gum unavailable.
// If gum is unavailable, defaults to proceeding (non-interactive environments).
func Confirm(prompt string) bool {
	if mode.IsAgent() || !IsAvailable() || !isTTY() {
		return true
	}
	cmd := exec.Command("gum", "confirm", prompt)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stderr // gum confirm renders on stderr
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil // exit 0 = confirmed, exit 1 = declined
}

// ConfirmDelete prints a styled warning and asks for confirmation.
func ConfirmDelete(resource string) bool {
	fmt.Fprintf(os.Stderr, "\n")
	return Confirm(fmt.Sprintf("Delete %s? This cannot be undone.", resource))
}
