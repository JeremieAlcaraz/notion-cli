package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AskPassword prompts for a masked/hidden input using gum input --password.
func AskPassword(prompt, placeholder string) (string, error) {
	if IsAvailable() && isTTY() {
		cmd := exec.Command("gum", "input",
			"--password",
			"--prompt", prompt,
			"--placeholder", placeholder,
		)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("input cancelled")
		}
		val := strings.TrimSpace(out.String())
		if val == "" {
			return "", fmt.Errorf("input cancelled")
		}
		return val, nil
	}
	return plainInput(prompt)
}

// AskChoose presents a list of options using gum choose and returns the selected one.
func AskChoose(header string, options []string) (string, error) {
	if !IsAvailable() || !isTTY() || len(options) == 0 {
		return "", fmt.Errorf("interactive selection not available")
	}
	args := append([]string{"choose", "--header", header}, options...)
	cmd := exec.Command("gum", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("selection cancelled")
	}
	return strings.TrimSpace(out.String()), nil
}

// AskInput prompts the user for a single value using gum input.
// Falls back to a plain fmt.Scan if gum is unavailable.
// Returns the trimmed input, or an error if the user cancelled.
func AskInput(prompt, placeholder string) (string, error) {
	if IsAvailable() && isTTY() {
		return gumInput(prompt, placeholder)
	}
	return plainInput(prompt)
}

func gumInput(prompt, placeholder string) (string, error) {
	args := []string{"input",
		"--prompt", prompt,
		"--placeholder", placeholder,
	}
	cmd := exec.Command("gum", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("input cancelled")
	}
	val := strings.TrimSpace(out.String())
	if val == "" {
		return "", fmt.Errorf("input cancelled")
	}
	return val, nil
}

func plainInput(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	var val string
	if _, err := fmt.Scan(&val); err != nil || val == "" {
		return "", fmt.Errorf("input cancelled")
	}
	return strings.TrimSpace(val), nil
}
