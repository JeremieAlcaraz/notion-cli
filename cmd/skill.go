package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed skill_content.md
var skillContent []byte

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage the notion-cli AI agent skill",
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the AI agent skill for Claude, Codex, or all detected agents",
	Long: `Install the notion-cli skill file for AI agents.

The skill is embedded in the binary and always matches the installed version.

Examples:
  notion skill install           # auto-detect and install for all found agents
  notion skill install --claude  # install for Claude (~/.claude/skills/notion-cli/)
  notion skill install --codex   # install for Codex (~/.codex/skills/notion-cli/)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		claude, _ := cmd.Flags().GetBool("claude")
		codex, _ := cmd.Flags().GetBool("codex")

		targets := map[string]string{}

		if claude {
			targets["Claude"] = filepath.Join(os.Getenv("HOME"), ".claude", "skills", "notion-cli")
		}
		if codex {
			targets["Codex"] = filepath.Join(os.Getenv("HOME"), ".codex", "skills", "notion-cli")
		}

		// Auto-detect if no flag specified
		if !claude && !codex {
			candidates := map[string]string{
				"Claude": filepath.Join(os.Getenv("HOME"), ".claude"),
				"Codex":  filepath.Join(os.Getenv("HOME"), ".codex"),
			}
			for agent, dir := range candidates {
				if _, err := os.Stat(dir); err == nil {
					targets[agent] = filepath.Join(dir, "skills", "notion-cli")
				}
			}
			if len(targets) == 0 {
				fmt.Println("No AI agent config directory found.")
				fmt.Println("Use --claude or --codex to install explicitly.")
				return nil
			}
		}

		for agent, dir := range targets {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			dest := filepath.Join(dir, "SKILL.md")
			if err := os.WriteFile(dest, skillContent, 0644); err != nil {
				return fmt.Errorf("failed to write skill file: %w", err)
			}
			fmt.Printf("✓ %s skill installed → %s\n", agent, dest)
		}
		return nil
	},
}

func init() {
	skillInstallCmd.Flags().Bool("claude", false, "Install for Claude (~/.claude/skills/notion-cli/)")
	skillInstallCmd.Flags().Bool("codex", false, "Install for Codex (~/.codex/skills/notion-cli/)")
	skillCmd.AddCommand(skillInstallCmd)
	rootCmd.AddCommand(skillCmd)
}
