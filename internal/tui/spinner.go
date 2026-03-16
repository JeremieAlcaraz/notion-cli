package tui

import (
	"fmt"
	"os"
	"time"
)

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// StartSpinner displays an animated spinner on stderr while work is in progress.
// Returns a stop function that clears the spinner line.
// No-op if gum is disabled or stderr is not a TTY.
func StartSpinner(title string) func() {
	if noGum || !isTTY() {
		return func() {}
	}

	done := make(chan struct{})

	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			case <-time.After(80 * time.Millisecond):
				fmt.Fprintf(os.Stderr, "\r%s %s ", spinnerFrames[i%len(spinnerFrames)], title)
				i++
			}
		}
	}()

	return func() {
		close(done)
		// Clear the spinner line
		fmt.Fprintf(os.Stderr, "\r\033[K")
	}
}

// isTTY reports whether stderr is a terminal.
func isTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
