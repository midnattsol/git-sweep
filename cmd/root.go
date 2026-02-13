package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/config"
	"github.com/midnattsol/git-sweep/internal/git"
	"github.com/midnattsol/git-sweep/internal/github"
	"github.com/midnattsol/git-sweep/internal/sweep"
	"github.com/midnattsol/git-sweep/internal/ui"
	"github.com/midnattsol/git-sweep/internal/update"
)

var (
	version = "dev"

	flagExecute bool
	flagNuke    bool
	flagYes     bool
	flagVerbose bool
	flagRemote  string
	flagNoColor bool
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-sweep",
		Short: "Safe local branch cleanup for squash-merge workflows",
		Long: `git-sweep safely deletes local branches that:
  1. Have an upstream that no longer exists (after git fetch --prune)
  2. Have a merged PR on GitHub

Use --nuke for interactive mode to delete any branch.`,
		Version:      version,
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&flagExecute, "execute", false, "Actually delete branches (default: dry-run)")
	cmd.Flags().BoolVar(&flagNuke, "nuke", false, "Interactive mode: select any branches to delete")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation in nuke mode (delete all non-protected)")
	cmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show skip reasons")
	cmd.Flags().StringVar(&flagRemote, "remote", "", "Remote name (default: origin, env: GIT_SWEEP_REMOTE)")
	cmd.Flags().BoolVar(&flagNoColor, "no-color", false, "Disable colors (env: GIT_SWEEP_NO_COLOR)")

	// Add subcommands
	cmd.AddCommand(NewUpdateCmd())

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

	// Create GitHub client
	var ghClient *github.Client
	var mergedPRs map[string]bool

	ms := ui.NewMultiSpinner()
	ms.Add(fmt.Sprintf("Fetching %s...", cfg.Remote), func() error {
		return git.FetchAndPrune(cfg.Remote)
	})
	ms.Add("Connecting to GitHub...", func() error {
		var err error
		ghClient, err = github.NewClient(cfg.Remote)
		return err
	})
	ms.Add("Loading merged PRs...", func() error {
		var err error
		ctx := context.Background()
		mergedPRs, err = ghClient.MergedPRBranches(ctx, cfg.Limit)
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

	// Show mode
	fmt.Print(ui.RenderMode(!flagExecute))

	// Execute if requested
	if flagExecute {
		sweep.Execute(result)
	}

	// Show results
	fmt.Print(ui.RenderBranchList(result, !flagExecute, flagVerbose))
	fmt.Print(ui.RenderSummary(result.Stats, !flagExecute))

	// Show tip if dry run and there are eligible branches
	if !flagExecute && result.Stats.Eligible > 0 {
		fmt.Print(ui.RenderTip())
	} else {
		fmt.Println()
	}

	// Check for updates in background if auto-update enabled
	CheckUpdateInBackground()

	return nil
}

func runNuke(cfg *config.Config) error {
	fmt.Print(ui.RenderHeader())

	// Fetch with spinner
	if err := ui.RunWithSpinner(fmt.Sprintf("Fetching %s...", cfg.Remote), func() error {
		return git.FetchAndPrune(cfg.Remote)
	}); err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Analyze for nuke
	result, err := sweep.AnalyzeForNuke(cfg)
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
		toDelete, err = ui.RunPicker(result)
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
