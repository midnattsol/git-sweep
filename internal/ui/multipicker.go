package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MultiPickerItem represents an item that can be displayed in the picker.
type MultiPickerItem interface {
	// Key returns the unique identifier for this item (used for selection tracking)
	Key() string
	// Label returns the display text for this item
	Label() string
	// Details returns optional additional info shown after the label (can be empty)
	Details() string
}

// MultiPickerGroup represents a group of items with an optional header.
type MultiPickerGroup struct {
	Header string
	Style  func(...string) string // Optional style function for the header (matches lipgloss.Style.Render)
	Items  []MultiPickerItem
}

// MultiPickerConfig configures the picker behavior.
type MultiPickerConfig struct {
	Title       string                               // Prompt shown at the top
	HelpText    string                               // Custom help text (defaults to standard)
	Groups      []MultiPickerGroup                   // Items grouped with headers
	PreSelected map[string]bool                      // Keys to pre-select
	ExtraKeys   map[string]func(m *MultiPickerModel) // Extra key handlers (e.g., "o" for old)
	ExtraHelp   string                               // Extra help text to append
}

// MultiPickerModel is a generic bubbletea model for item selection.
type MultiPickerModel struct {
	config    MultiPickerConfig
	items     []MultiPickerItem // Flattened list of all items
	selected  map[string]bool
	cursor    int
	quitting  bool
	confirmed bool
}

// NewMultiPicker creates a new picker model with the given configuration.
func NewMultiPicker(cfg MultiPickerConfig) MultiPickerModel {
	// Flatten items from groups
	var items []MultiPickerItem
	for _, g := range cfg.Groups {
		items = append(items, g.Items...)
	}

	// Initialize selection
	selected := make(map[string]bool)
	if cfg.PreSelected != nil {
		for k, v := range cfg.PreSelected {
			selected[k] = v
		}
	}

	return MultiPickerModel{
		config:   cfg,
		items:    items,
		selected: selected,
	}
}

func (m MultiPickerModel) Init() tea.Cmd {
	return nil
}

func (m MultiPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Check extra key handlers first
		if m.config.ExtraKeys != nil {
			if handler, ok := m.config.ExtraKeys[key]; ok {
				handler(&m)
				return m, nil
			}
		}

		switch key {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.confirmed = true
			return m, tea.Quit

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}

		case " ":
			// Toggle selection
			if m.cursor < len(m.items) {
				key := m.items[m.cursor].Key()
				m.selected[key] = !m.selected[key]
			}

		case "a":
			// Select all
			for _, item := range m.items {
				m.selected[item.Key()] = true
			}

		case "n":
			// Select none
			m.selected = make(map[string]bool)
		}
	}

	return m, nil
}

func (m MultiPickerModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(fmt.Sprintf("\n  %s\n", MutedStyle.Render(m.config.Title)))

	// Render groups with headers
	itemIdx := 0
	for _, group := range m.config.Groups {
		if len(group.Items) == 0 {
			continue
		}

		// Group header
		if group.Header != "" {
			header := group.Header
			if group.Style != nil {
				header = group.Style(header)
			}
			b.WriteString(fmt.Sprintf("\n  %s\n", header))
		} else {
			b.WriteString("\n")
		}

		// Group items
		for _, item := range group.Items {
			b.WriteString(m.renderItem(item, itemIdx))
			itemIdx++
		}
	}

	// Help
	helpText := m.config.HelpText
	if helpText == "" {
		helpText = "␣ toggle · a all · n none · ↵ confirm · q quit"
		if m.config.ExtraHelp != "" {
			helpText = m.config.ExtraHelp + " · " + helpText
		}
	}

	b.WriteString(fmt.Sprintf("\n  %s\n", Divider(len(helpText)+4)))
	b.WriteString(fmt.Sprintf("  %s\n\n", HelpStyle.Render(helpText)))

	return b.String()
}

func (m MultiPickerModel) renderItem(item MultiPickerItem, idx int) string {
	cursor := "  "
	if idx == m.cursor {
		cursor = CursorStyle.Render() + " "
	}

	var checkbox string
	if m.selected[item.Key()] {
		checkbox = SuccessStyle.Render("▣")
	} else {
		checkbox = "▢"
	}

	label := item.Label()
	if idx == m.cursor {
		label = SelectedStyle.Render(label)
	}

	details := item.Details()
	if details != "" {
		details = "  " + MutedStyle.Render(details)
	}

	return fmt.Sprintf("%s%s %s%s\n", cursor, checkbox, label, details)
}

// Cancelled returns true if the user quit without confirming.
func (m MultiPickerModel) Cancelled() bool {
	return m.quitting
}

// Selected returns the keys of all selected items.
func (m MultiPickerModel) Selected() []string {
	var keys []string
	for key, sel := range m.selected {
		if sel {
			keys = append(keys, key)
		}
	}
	return keys
}

// IsSelected returns true if the item with the given key is selected.
func (m MultiPickerModel) IsSelected(key string) bool {
	return m.selected[key]
}

// SetSelected sets the selection state for a key.
func (m *MultiPickerModel) SetSelected(key string, selected bool) {
	m.selected[key] = selected
}

// ClearSelection clears all selections.
func (m *MultiPickerModel) ClearSelection() {
	m.selected = make(map[string]bool)
}

// Items returns all items in the picker.
func (m MultiPickerModel) Items() []MultiPickerItem {
	return m.items
}

// RunMultiPicker runs the picker and returns selected keys, or nil if cancelled.
func RunMultiPicker(cfg MultiPickerConfig) ([]string, error) {
	m := NewMultiPicker(cfg)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(MultiPickerModel)
	if fm.Cancelled() {
		return nil, nil
	}

	return fm.Selected(), nil
}
