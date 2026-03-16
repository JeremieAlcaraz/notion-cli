package cmd

import (
	"fmt"
	"os"

	"github.com/4ier/notion-cli/cmd/generated"
	"github.com/4ier/notion-cli/internal/config"
	"github.com/4ier/notion-cli/internal/tui"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	fieldFilter  string
	debugMode    bool
	dryRunFlag   bool
	noGumFlag    bool
	agentFlag    bool
	// Version is set by goreleaser ldflags
	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "notion",
	Short:   "Work seamlessly with Notion from the command line",
	Long: `Work seamlessly with Notion from the command line.

Notion CLI lets you manage pages, databases, blocks, and more
without leaving your terminal. Built for developers and AI agents.`,
	Version:          Version,
	SilenceUsage:     true,
	SilenceErrors:    true,
	TraverseChildren: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "", "Output format: json, md, table, text (default: auto)")
	rootCmd.PersistentFlags().StringVar(&fieldFilter, "field", "", "Extract a single top-level field from the JSON response")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Show HTTP request/response details")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Print the HTTP request without executing it")
	rootCmd.PersistentFlags().BoolVar(&noGumFlag, "no-gum", false, "Disable gum TUI enhancements (plain text output)")
	rootCmd.PersistentFlags().BoolVar(&agentFlag, "agent", false, "Agent mode: minified JSON, no TUI, terse errors (also via NOTION_AGENT=1)")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		tui.InitAgentMode()         // check NOTION_AGENT env first
		tui.SetAgentMode(agentFlag) // --agent flag overrides
		tui.SetNoGum(noGumFlag)
		generated.SetDryRun(dryRunFlag)
		tui.WarnIfMissing()
	}

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(apiCmd)

	// Generated commands from OpenAPI spec
	generated.AddTo(rootCmd)
}

// getToken returns the Notion API token from flag, env, or config file.
func getToken() (string, error) {
	// 1. Environment variable
	if token := os.Getenv("NOTION_TOKEN"); token != "" {
		return token, nil
	}

	// 2. Config file (with profile support)
	cfg, err := config.Load()
	if err == nil {
		profile := cfg.GetCurrentProfile()
		if profile != nil && profile.Token != "" {
			return profile.Token, nil
		}
	}

	return "", fmt.Errorf("not authenticated. Run 'notion auth login --with-token' or set NOTION_TOKEN")
}
