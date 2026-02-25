package spinner

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	spinner  spinner.Model
	spinning bool
	message  string
	style    lipgloss.Style
}

func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"", "", "", "", "", ""},
		FPS:    time.Second / 10,
	}
	s.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	return Model{
		spinner: s,
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true),
	}
}

func (m *Model) Start(message string) tea.Cmd {
	m.spinning = true
	m.message = message
	return func() tea.Msg { return m.spinner.Tick() }
}

func (m *Model) Stop() {
	m.spinning = false
	m.message = ""
}

func (m *Model) IsSpinning() bool {
	return m.spinning
}

func (m *Model) Message() string {
	return m.message
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.spinning {
		return *m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return *m, cmd
}

func (m Model) View() string {
	if !m.spinning {
		return ""
	}

	return m.spinner.View() + " " + m.style.Render(m.message)
}
