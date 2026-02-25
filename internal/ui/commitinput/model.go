package commitinput

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model represents the inline commit message input
type Model struct {
	input   textinput.Model
	width   int
	focused bool

	containerStyle lipgloss.Style
	labelStyle     lipgloss.Style
}

// New creates a new commit input model
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.CharLimit = 200
	ti.SetWidth(40)
	ti.Prompt = " "

	placeholderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("78")).
		Bold(true)

	s := ti.Styles()
	s.Focused.Placeholder = placeholderStyle
	s.Blurred.Placeholder = placeholderStyle
	s.Focused.Prompt = promptStyle
	s.Blurred.Prompt = promptStyle
	ti.SetStyles(s)

	return Model{
		input: ti,
		containerStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1),
		labelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")).
			Bold(true),
	}
}

// SetWidth updates the width
func (m *Model) SetWidth(width int) {
	m.width = width
	m.input.SetWidth(width - 6) // Account for padding/borders
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

	var label string
	if m.focused {
		focusedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1")).
			Bold(true)
		label = focusedStyle.Render(" Commit")
	} else {
		label = m.labelStyle.Render(" Commit")
	}

	containerStyle := m.containerStyle
	if m.focused {
		containerStyle = containerStyle.
			BorderForeground(lipgloss.Color("#a6e3a1"))
	}

	content := label + "\n" + m.input.View()

	return containerStyle.Width(m.width - 2).Render(content)
}

// Height returns the height of the component
func (m Model) Height() int {
	return 4 // Label + input + border
}
