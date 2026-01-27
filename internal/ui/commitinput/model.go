package commitinput

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the inline commit message input
type Model struct {
	input   textinput.Model
	width   int
	focused bool

	// Styles
	containerStyle lipgloss.Style
	labelStyle     lipgloss.Style
	inputStyle     lipgloss.Style
	placeholderStyle lipgloss.Style
}

// New creates a new commit input model
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter commit message..."
	ti.CharLimit = 200
	ti.Width = 40
	// Style the placeholder
	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	// Style the text cursor
	ti.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))
	// Style the prompt
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("78")).
		Bold(true)
	ti.Prompt = " "

	return Model{
		input: ti,
		containerStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1),
		labelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")).
			Bold(true),
		inputStyle:       lipgloss.NewStyle(),
		placeholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true),
	}
}

// SetWidth updates the width
func (m *Model) SetWidth(width int) {
	m.width = width
	m.input.Width = width - 6 // Account for padding/borders
}

// Focus focuses the input
func (m *Model) Focus() tea.Cmd {
	m.focused = true
	return m.input.Focus()
}

// Blur removes focus from the input
func (m *Model) Blur() {
	m.focused = false
	m.input.Blur()
}

// IsFocused returns whether the input is focused
func (m Model) IsFocused() bool {
	return m.focused
}

// Value returns the current input value
func (m Model) Value() string {
	return m.input.Value()
}

// SetValue sets the input value
func (m *Model) SetValue(v string) {
	m.input.SetValue(v)
}

// Reset clears the input
func (m *Model) Reset() {
	m.input.Reset()
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Commit icon and label with focus indicator
	icon := "" // Git commit icon
	var label string
	if m.focused {
		// Highlight when focused
		focusedIcon := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Render(icon)
		focusedLabel := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Render(" Commit")
		label = focusedIcon + focusedLabel
	} else {
		label = m.labelStyle.Render(icon + " Commit")
	}

	b.WriteString(label)
	b.WriteString("\n")

	// Input field
	inputView := m.input.View()

	// Apply container style with dynamic border color based on focus
	containerStyle := m.containerStyle
	if m.focused {
		containerStyle = containerStyle.
			BorderForeground(lipgloss.Color("214")).
			BorderStyle(lipgloss.RoundedBorder())
	}

	content := b.String() + inputView

	return containerStyle.Width(m.width - 2).Render(content)
}

// Height returns the height of the component
func (m Model) Height() int {
	return 4 // Label + input + border
}
