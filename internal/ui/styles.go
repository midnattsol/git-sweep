package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Soft color palette - easier on the eyes
var (
	// Primary colors (softer tones)
	Purple    = lipgloss.Color("#C084FC") // lavanda/magenta suave
	Pink      = lipgloss.Color("#F0ABFC") // pink suave
	Cyan      = lipgloss.Color("#22D3EE") // cyan claro
	Green     = lipgloss.Color("#34D399") // verde menta
	Yellow    = lipgloss.Color("#FBBF24") // amarillo suave
	Red       = lipgloss.Color("#F87171") // rojo suave
	Gray      = lipgloss.Color("#9CA3AF")
	DarkGray  = lipgloss.Color("#4B5563")
	LightGray = lipgloss.Color("#D1D5DB")
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
