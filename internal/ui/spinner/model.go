package spinner

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model wraps the bubbles spinner for use in the app
type Model struct {
	spinner  spinner.Model
	spinning bool
	message  string
	style    lipgloss.Style
}

// New creates a new spinner model
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		spinner: s,
		style:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
	}
}

// Start begins the spinner with a message
func (m *Model) Start(message string) tea.Cmd {
	m.spinning = true
	m.message = message
	return m.spinner.Tick
}

// Stop stops the spinner
func (m *Model) Stop() {
	m.spinning = false
	m.message = ""
}

// IsSpinning returns whether the spinner is active
func (m *Model) IsSpinning() bool {
	return m.spinning
}

// Message returns the current spinner message
func (m *Model) Message() string {
	return m.message
}

// Update handles spinner tick messages
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.spinning {
		return *m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return *m, cmd
}

// View renders the spinner with its message
func (m Model) View() string {
	if !m.spinning {
		return ""
	}

	return m.spinner.View() + " " + m.style.Render(m.message)
}
