package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Charm-style color palette
var (
	// Primary colors
	Purple    = lipgloss.Color("#7C3AED")
	Pink      = lipgloss.Color("#EC4899")
	Cyan      = lipgloss.Color("#06B6D4")
	Green     = lipgloss.Color("#10B981")
	Yellow    = lipgloss.Color("#F59E0B")
	Red       = lipgloss.Color("#EF4444")
	Gray      = lipgloss.Color("#6B7280")
	DarkGray  = lipgloss.Color("#374151")
	LightGray = lipgloss.Color("#9CA3AF")
)

// Styles
var (
	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Purple)

	// Subtitle / info
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(LightGray)

	// Success
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green)

	// Warning
	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	// Error
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red)

	// Muted / dim text
	MutedStyle = lipgloss.NewStyle().
			Foreground(Gray)

	// Bold
	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	// Branch name
	BranchStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	// Checkmark
	CheckStyle = lipgloss.NewStyle().
			Foreground(Green).
			SetString("✓")

	// Cross
	CrossStyle = lipgloss.NewStyle().
			Foreground(Red).
			SetString("✗")

	// Circle (skipped)
	CircleStyle = lipgloss.NewStyle().
			Foreground(Gray).
			SetString("◦")

	// Box styles for header/footer
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DarkGray).
			Padding(0, 1)

	// Divider
	DividerStyle = lipgloss.NewStyle().
			Foreground(DarkGray)

	// Selected item in list
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true)

	// Cursor
	CursorStyle = lipgloss.NewStyle().
			Foreground(Pink).
			SetString("›")

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(LightGray)

	// Mode indicator
	DryRunStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	ExecuteStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)
)

// Divider returns a horizontal divider line
func Divider(width int) string {
	return DividerStyle.Render(repeat("─", width))
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
