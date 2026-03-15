// Package generated contains CLI commands generated from the Notion OpenAPI spec.
// Do not edit generated files — run `just generate` to regenerate.
// This file (helpers.go) is NOT generated and may be edited manually.
package generated

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/4ier/notion-cli/internal/config"
)

// resolveBody resolves the --body flag value.
//   - "--body -"         → reads from stdin
//   - "--body @file.json" → reads from file
//   - "--body {...}"     → uses the value as-is
func resolveBody(s string) (string, error) {
	if s == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	if strings.HasPrefix(s, "@") {
		path := s[1:]
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read file %s: %w", path, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return s, nil
}

// GetToken returns the active Notion API token from env or config file.
func GetToken() (string, error) {
	if token := os.Getenv("NOTION_TOKEN"); token != "" {
		return token, nil
	}
	cfg, err := config.Load()
	if err == nil {
		if profile := cfg.GetCurrentProfile(); profile != nil && profile.Token != "" {
			return profile.Token, nil
		}
	}
	return "", fmt.Errorf("not authenticated: run 'notion auth login --with-token' or set NOTION_TOKEN")
}
