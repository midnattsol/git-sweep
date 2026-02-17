package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/config"
	"github.com/midnattsol/git-sweep/internal/git"
	"github.com/midnattsol/git-sweep/internal/provider"
	"github.com/midnattsol/git-sweep/internal/sweep"
	"github.com/midnattsol/git-sweep/internal/ui"
	"github.com/midnattsol/git-sweep/internal/update"
)

var (
	version = "dev"

	flagYes     bool
	flagRemote  string
	flagNoColor bool
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-sweep",
		Short: "Interactive local branch cleanup",
		Long: `git-sweep analyzes local branches and opens an interactive picker.

Suggested branches are pre-selected (merged PR, gone upstream, stale with no upstream).
Use --yes to skip interaction and delete all suggested branches.`,
		Version:      version,
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip interaction and delete all suggested branches")
	cmd.Flags().StringVar(&flagRemote, "remote", "", "Remote name (default: origin, env: GIT_SWEEP_REMOTE)")
	cmd.Flags().BoolVar(&flagNoColor, "no-color", false, "Disable colors (env: GIT_SWEEP_NO_COLOR)")

	// Add subcommands
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewStashCmd())
	cmd.AddCommand(NewTagsCmd())
	cmd.AddCommand(NewWorktreeCmd())
	cmd.AddCommand(NewProtectCmd())
	cmd.AddCommand(NewUnprotectCmd())

	return cmd
}

func Execute() {
	// Set version for update package
	update.CurrentVersion = version

	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Load config
	cfg := config.Load()

	// Override with flags
	if flagRemote != "" {
		cfg.Remote = flagRemote
	}
	if flagNoColor {
		cfg.NoColor = true
	}

	// Preconditions
	if !git.IsInsideWorkTree() {
		fmt.Print(ui.RenderError("Not a git repository"))
		return fmt.Errorf("not a git repository")
	}

	return runBranchCleanup(cfg)
}

func runBranchCleanup(cfg *config.Config) error {
	fmt.Print(ui.RenderNukeHeader())

	// Variables for provider and merged PRs
	var prov provider.Provider
	var mergedPRs map[string]bool
	var providerErr error

	ms := ui.NewMultiSpinner()
	ms.Add(fmt.Sprintf("Fetching %s...", cfg.Remote), func() error {
		return git.FetchAndPrune(cfg.Remote)
	})
	ms.Add("Detecting provider...", func() error {
		info, err := provider.DetectFromRemote(cfg.Remote)
		if err != nil {
			providerErr = err
			return nil // Don't fail, cleanup can work without provider
		}
		provCfg := &provider.Config{
			GitHubToken:       cfg.GitHubToken,
			GitLabToken:       cfg.GitLabToken,
			GitLabURL:         cfg.GitLabURL,
			BitbucketUsername: cfg.BitbucketUsername,
			BitbucketToken:    cfg.BitbucketToken,
		}
		prov, err = provider.New(info, provCfg)
		if err != nil {
			providerErr = err
			return nil // Don't fail, cleanup can work without provider
		}
		return nil
	})
	ms.Add("Loading merged PRs...", func() error {
		if prov == nil {
			return nil // Skip if no provider
		}
		ctx := context.Background()
		var err error
		mergedPRs, err = prov.MergedPRBranches(ctx, cfg.Limit)
		if err != nil {
			providerErr = err
			return nil // Don't fail, just won't have PR info
		}
		return nil
	})

	if err := ms.Run(); err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Get current branch and default branch for potential checkout
	currentBranch, _ := git.CurrentBranch()
	defaultBranch, _ := git.DefaultBranch(cfg.Remote)

	// Analyze branches (with or without merged PRs)
	result, err := sweep.AnalyzeBranches(cfg, mergedPRs)
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Check if there are any branches to delete
	if result.Stats.Eligible == 0 {
		fmt.Print(ui.RenderNoBranches())
		CheckUpdateInBackground()
		return nil
	}

	var toDelete []string

	if flagYes {
		// No interaction, delete only suggested branches
		// Note: current branch is never pre-selected as suggested
		for _, br := range result.Branches {
			if br.Suggested && !br.IsCurrent {
				toDelete = append(toDelete, br.Branch.Name)
			}
		}
	} else {
		if !ui.IsTTY() {
			err := fmt.Errorf("interactive mode requires a terminal; use --yes to delete suggested branches")
			fmt.Print(ui.RenderError(err.Error()))
			return err
		}

		// Interactive picker (passes defaultBranch for current branch confirmation)
		toDelete, err = ui.RunPicker(result, providerErr, defaultBranch)
		if err != nil {
			fmt.Print(ui.RenderError(err.Error()))
			return err
		}

		if toDelete == nil {
			// User cancelled
			return nil
		}
	}

	if len(toDelete) == 0 {
		fmt.Print(ui.RenderNoBranches())
		return nil
	}

	// Delete selected branches
	deleted, errors := sweep.DeleteBranches(toDelete, currentBranch, defaultBranch)

	// Show errors if any
	for _, err := range errors {
		fmt.Printf("  %s %s\n", ui.CrossStyle.Render(), ui.ErrorStyle.Render(err.Error()))
	}

	fmt.Print(ui.RenderNukeSummary(deleted, len(toDelete)))

	// Check for updates in background if auto-update enabled
	CheckUpdateInBackground()

	return nil
}
