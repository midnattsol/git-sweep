package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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

	flagDryRun     bool
	flagForce      bool
	flagOnlyMerged bool
	flagNuke       bool
	flagYes        bool
	flagBrief      bool
	flagRemote     string
	flagNoColor    bool
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-sweep",
		Short: "Safe local branch cleanup for squash-merge workflows",
		Long: `git-sweep safely deletes local branches with gone upstream.

By default, deletes all branches where the upstream no longer exists.
Use --only-merged to require a confirmed merged PR before deletion.
Use --nuke for interactive mode to delete any branch.`,
		Version:      version,
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Only show what would be deleted, don't prompt or delete")
	cmd.Flags().BoolVar(&flagForce, "force", false, "Delete without confirmation")
	cmd.Flags().BoolVar(&flagOnlyMerged, "only-merged", false, "Only delete branches with confirmed merged PR")
	cmd.Flags().BoolVar(&flagNuke, "nuke", false, "Interactive mode: select any branches to delete")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation in nuke mode (delete all suggested)")
	cmd.Flags().BoolVarP(&flagBrief, "brief", "b", false, "Hide skipped branches")
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

	if flagNuke {
		return runNuke(cfg)
	}

	return runSafe(cfg)
}

func runSafe(cfg *config.Config) error {
	fmt.Print(ui.RenderHeader())

	// Detect provider
	var prov provider.Provider
	var mergedPRs map[string]bool

	ms := ui.NewMultiSpinner()
	ms.Add(fmt.Sprintf("Fetching %s...", cfg.Remote), func() error {
		return git.FetchAndPrune(cfg.Remote)
	})
	ms.Add("Detecting provider...", func() error {
		info, err := provider.DetectFromRemote(cfg.Remote)
		if err != nil {
			return err
		}
		provCfg := &provider.Config{
			GitHubToken:       cfg.GitHubToken,
			GitLabToken:       cfg.GitLabToken,
			GitLabURL:         cfg.GitLabURL,
			BitbucketUsername: cfg.BitbucketUsername,
			BitbucketToken:    cfg.BitbucketToken,
		}
		prov, err = provider.New(info, provCfg)
		return err
	})
	ms.Add("Loading merged PRs...", func() error {
		var err error
		ctx := context.Background()
		mergedPRs, err = prov.MergedPRBranches(ctx, cfg.Limit)
		return err
	})

	if err := ms.Run(); err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Analyze
	result, err := sweep.Analyze(cfg, mergedPRs)
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Show results
	fmt.Print(ui.RenderBranchList(result, !flagOnlyMerged, !flagBrief))
	fmt.Print(ui.RenderSummary(result.Stats, !flagOnlyMerged))

	// Count branches to delete
	toDeleteCount := sweep.CountToDelete(result, !flagOnlyMerged)

	// If nothing to delete, we're done
	if toDeleteCount == 0 {
		fmt.Print(ui.RenderNoBranches())
		CheckUpdateInBackground()
		return nil
	}

	// Dry run mode: just show, don't prompt or delete
	if flagDryRun {
		fmt.Print(ui.RenderDryRunTip())
		CheckUpdateInBackground()
		return nil
	}

	// Ask for confirmation unless --force
	if !flagForce {
		fmt.Printf("\n  Delete %d branches? [y/N] ", toDeleteCount)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Printf("\n  %s Cancelled\n\n", ui.MutedStyle.Render("●"))
			return nil
		}
	}

	// Execute deletion
	sweep.Execute(result, !flagOnlyMerged)

	// Show deletion results
	fmt.Print(ui.RenderDeletedBranches(result))
	fmt.Print(ui.RenderDeleteSummary(result.Stats.Deleted, toDeleteCount))

	// Check for updates in background if auto-update enabled
	CheckUpdateInBackground()

	return nil
}

func runNuke(cfg *config.Config) error {
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
			return nil // Don't fail, nuke can work without provider
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
			return nil // Don't fail, nuke can work without provider
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

	// Analyze for nuke (with or without merged PRs)
	result, err := sweep.AnalyzeForNuke(cfg, mergedPRs)
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Check if there are any branches to delete
	if result.Stats.Eligible == 0 {
		fmt.Print(ui.RenderNoBranches())
		return nil
	}

	var toDelete []string

	if flagYes {
		// No interaction, delete all eligible
		for _, br := range result.Branches {
			if br.Eligible {
				toDelete = append(toDelete, br.Branch.Name)
			}
		}
	} else {
		// Interactive picker
		toDelete, err = ui.RunPicker(result, providerErr)
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
	deleted, errors := sweep.DeleteBranches(toDelete)

	// Show results
	for _, name := range toDelete {
		fmt.Printf("  %s %s\n", ui.CheckStyle.Render(), ui.BranchStyle.Render(name))
	}

	for _, err := range errors {
		fmt.Printf("  %s %s\n", ui.CrossStyle.Render(), ui.ErrorStyle.Render(err.Error()))
	}

	fmt.Print(ui.RenderNukeSummary(deleted, len(toDelete)))

	// Check for updates in background if auto-update enabled
	CheckUpdateInBackground()

	return nil
}
