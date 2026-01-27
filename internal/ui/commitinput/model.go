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
	ti.Placeholder = "Message (press Enter to commit)"
	ti.CharLimit = 200
	ti.Width = 40

	return Model{
		input:            ti,
		containerStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")).Padding(0, 1),
		labelStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Bold(true),
		inputStyle:       lipgloss.NewStyle(),
		placeholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
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

	// Commit icon and label
	icon := "" // Git commit icon
	label := m.labelStyle.Render(icon + " Commit")

	b.WriteString(label)
	b.WriteString("\n")

	// Input field
	inputView := m.input.View()

	// Apply container style
	content := b.String() + inputView

	return m.containerStyle.Width(m.width - 2).Render(content)
}

// Height returns the height of the component
func (m Model) Height() int {
	return 4 // Label + input + border
}
