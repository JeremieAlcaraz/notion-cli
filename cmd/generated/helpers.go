// Package generated contains CLI commands generated from the Notion OpenAPI spec.
// Do not edit generated files — run `just generate` to regenerate.
// This file (helpers.go) is NOT generated and may be edited manually.
package generated

import (
	"fmt"
	"os"

	"github.com/4ier/notion-cli/internal/config"
)

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
