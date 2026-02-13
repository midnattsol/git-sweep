package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/config"
	"github.com/midnattsol/git-sweep/internal/ui"
)

func NewUnprotectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unprotect [pattern]",
		Short: "Remove a branch pattern from protected list",
		Long: `Remove a branch pattern from the protected list.

Note: Default patterns (master, main, develop) cannot be removed.
Use GIT_SWEEP_PROTECTED env var to override defaults completely.

Examples:
  git sweep unprotect staging         # Remove 'staging' from protected
  git sweep unprotect "release/**"    # Remove release pattern`,
		RunE: runUnprotect,
	}

	return cmd
}

func runUnprotect(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("specify a branch pattern to unprotect")
	}

	pattern := args[0]

	// Check if it's a default pattern
	defaults := map[string]bool{"master": true, "main": true, "develop": true}
	if defaults[pattern] {
		fmt.Printf("  %s %s is a default pattern and cannot be removed\n", ui.WarningStyle.Render("!"), ui.BranchStyle.Render(pattern))
		fmt.Println("  Use GIT_SWEEP_PROTECTED env var to override defaults")
		return nil
	}

	// Load current config
	fc, err := config.LoadFileConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Remove pattern
	if !fc.RemoveProtected(pattern) {
		fmt.Printf("  %s is not in the protected list\n", ui.BranchStyle.Render(pattern))
		return nil
	}

	// Save config
	if err := config.SaveFileConfig(fc); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("  %s Removed %s from protected branches\n", ui.SuccessStyle.Render("✓"), ui.BranchStyle.Render(pattern))
	return nil
}
