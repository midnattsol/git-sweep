package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/midnattsol/git-sweep/internal/ui"
	"github.com/midnattsol/git-sweep/internal/update"
)

var (
	flagCheck      bool
	flagYesUpdate  bool
	flagAutoUpdate string
)

func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update git-sweep to the latest version",
		Long: `Check for and install updates to git-sweep.

Examples:
  git-sweep update              # Check and prompt to update
  git-sweep update --check      # Only check, don't install
  git-sweep update --yes        # Update without confirmation
  git-sweep update --auto-update=true   # Enable auto-update
  git-sweep update --auto-update=false  # Disable auto-update`,
		RunE: runUpdate,
	}

	cmd.Flags().BoolVar(&flagCheck, "check", false, "Only check for updates, don't install")
	cmd.Flags().BoolVar(&flagYesUpdate, "yes", false, "Update without confirmation")
	cmd.Flags().StringVar(&flagAutoUpdate, "auto-update", "", "Enable/disable auto-update (true/false)")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Handle --auto-update flag
	if flagAutoUpdate != "" {
		enabled := strings.ToLower(flagAutoUpdate) == "true"
		if err := update.SetAutoUpdate(enabled); err != nil {
			fmt.Print(ui.RenderError(fmt.Sprintf("Failed to save config: %v", err)))
			return err
		}
		if enabled {
			fmt.Printf("  %s Auto-update enabled\n\n", ui.CheckStyle.Render())
		} else {
			fmt.Printf("  %s Auto-update disabled\n\n", ui.CheckStyle.Render())
		}
		return nil
	}

	// Show current version
	fmt.Printf("\n  %s Current version: %s\n", ui.CheckStyle.Render(), ui.BoldStyle.Render(update.CurrentVersion))

	// Check for updates with spinner
	var release *update.Release
	var hasUpdate bool

	if err := ui.RunWithSpinner("Checking for updates...", func() error {
		var err error
		release, hasUpdate, err = update.CheckForUpdate(ctx)
		return err
	}); err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	if !hasUpdate {
		fmt.Printf("\n  %s You're on the latest version\n\n", ui.CheckStyle.Render())
		return nil
	}

	// Show available update
	fmt.Printf("\n  %s New version available: %s\n\n",
		ui.WarningStyle.Render("●"),
		ui.SuccessStyle.Render(release.TagName))

	// If only checking, exit here
	if flagCheck {
		fmt.Printf("  Run %s to update.\n\n", ui.BoldStyle.Render("git sweep update"))
		return nil
	}

	// Ask for confirmation unless --yes
	if !flagYesUpdate {
		fmt.Print("  Do you want to update? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Printf("\n  %s Update cancelled\n\n", ui.MutedStyle.Render("●"))
			return nil
		}
		fmt.Println()
	}

	// Get download URL for current platform
	downloadURL, err := release.GetAssetForPlatform()
	if err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	// Download and install
	if err := ui.RunWithSpinner(fmt.Sprintf("Downloading %s...", release.TagName), func() error {
		return update.DownloadAndInstall(ctx, downloadURL)
	}); err != nil {
		fmt.Print(ui.RenderError(err.Error()))
		return err
	}

	fmt.Printf("\n  %s Updated to %s\n\n", ui.CheckStyle.Render(), ui.SuccessStyle.Render(release.TagName))

	return nil
}

// CheckUpdateInBackground checks for updates and prints a notice if available
// Used when auto-update is enabled
func CheckUpdateInBackground() {
	if !update.IsAutoUpdateEnabled() {
		return
	}

	ctx := context.Background()
	release, hasUpdate, err := update.CheckForUpdate(ctx)
	if err != nil || !hasUpdate {
		return
	}

	fmt.Printf("\n  %s New version %s available. Run %s to upgrade.\n",
		ui.MutedStyle.Render("Tip:"),
		ui.SuccessStyle.Render(release.TagName),
		ui.BoldStyle.Render("git sweep update"))
}
