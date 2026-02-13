package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/config"
	"github.com/midnattsol/git-sweep/internal/ui"
)

var flagProtectList bool

func NewProtectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "protect [pattern]",
		Short: "Add a branch pattern to protected list",
		Long: `Add a branch pattern to the protected list.

Protected branches are never deleted by git-sweep.

Patterns support wildcards:
  staging        - exact match
  release/*      - matches release/1.0 but not release/v2/hotfix  
  release/**     - matches release/1.0 and release/v2/hotfix

Examples:
  git sweep protect staging           # Protect 'staging' branch
  git sweep protect "release/**"      # Protect all release branches
  git sweep protect --list            # Show all protected patterns`,
		RunE: runProtect,
	}

	cmd.Flags().BoolVar(&flagProtectList, "list", false, "List all protected patterns")

	return cmd
}

func runProtect(cmd *cobra.Command, args []string) error {
	if flagProtectList {
		return listProtected()
	}

	if len(args) == 0 {
		return fmt.Errorf("specify a branch pattern to protect, or use --list")
	}

	pattern := args[0]

	// Load current config
	fc, err := config.LoadFileConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Add pattern
	if !fc.AddProtected(pattern) {
		fmt.Printf("  %s is already protected\n", ui.BranchStyle.Render(pattern))
		return nil
	}

	// Save config
	if err := config.SaveFileConfig(fc); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("  %s Added %s to protected branches\n", ui.SuccessStyle.Render("✓"), ui.BranchStyle.Render(pattern))
	return nil
}

func listProtected() error {
	patterns, err := config.GetProtectedPatterns()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println()
	fmt.Println("  Protected branch patterns:")
	fmt.Println()

	// Show defaults
	defaults := map[string]bool{"master": true, "main": true, "develop": true}

	for _, p := range patterns {
		if defaults[p] {
			fmt.Printf("    %s %s\n", ui.BranchStyle.Render(p), ui.MutedStyle.Render("(default)"))
		} else {
			fmt.Printf("    %s\n", ui.BranchStyle.Render(p))
		}
	}
	fmt.Println()

	return nil
}
