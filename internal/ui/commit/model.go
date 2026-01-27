package commit

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Danny-Dasilva/gdiff/internal/git"
	"github.com/Danny-Dasilva/gdiff/internal/types"
)

// Model represents the commit modal
type Model struct {
	textarea textarea.Model
	width    int
	height   int
	visible  bool
	amend    bool
	keyMap   types.KeyMap

	// Editor support
	tempFilePath string

	// Styles
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
	helpStyle   lipgloss.Style
}

// New creates a new commit modal
func New(keyMap types.KeyMap) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter commit message..."
	ta.Focus()
	ta.SetWidth(60)
	ta.SetHeight(10)
	ta.CharLimit = 0 // No limit

	return Model{
		textarea:    ta,
		keyMap:      keyMap,
		borderStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1),
		titleStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
		helpStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}

// Show displays the commit modal
func (m *Model) Show(amend bool) {
	m.visible = true
	m.amend = amend
	m.textarea.Reset()
	m.textarea.Focus()
}

// Hide hides the commit modal
func (m *Model) Hide() {
	m.visible = false
	m.textarea.Blur()
}

// Visible returns whether the modal is visible
func (m Model) Visible() bool {
	return m.visible
}

// Message returns the commit message
func (m Model) Message() string {
	return strings.TrimSpace(m.textarea.Value())
}

// Amend returns whether this is an amend commit
func (m Model) Amend() bool {
	return m.amend
}

// SetSize updates the modal size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Modal is 70% of terminal width, max 80 chars
	modalWidth := width * 70 / 100
	if modalWidth > 80 {
		modalWidth = 80
	}
	if modalWidth < 40 {
		modalWidth = 40
	}

	m.textarea.SetWidth(modalWidth - 4) // Account for borders
	m.textarea.SetHeight(10)
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// ConfirmMsg is sent when the user confirms the commit
type ConfirmMsg struct {
	Message string
	Amend   bool
}

// CancelMsg is sent when the user cancels the commit
type CancelMsg struct{}

// EditorFinishedMsg is sent when the external editor closes
type EditorFinishedMsg struct {
	Content string
	Err     error
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case EditorFinishedMsg:
		// Editor closed, update textarea with content
		if msg.Err == nil && msg.Content != "" {
			m.textarea.SetValue(msg.Content)
		}
		// Clean up temp file
		if m.tempFilePath != "" {
			os.Remove(m.tempFilePath)
			m.tempFilePath = ""
		}
		m.textarea.Focus()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Escape):
			m.Hide()
			return m, func() tea.Msg { return CancelMsg{} }

		case key.Matches(msg, m.keyMap.OpenEditor):
			// Ctrl+E to open in external editor
			return m, m.openEditor()

		case msg.Type == tea.KeyCtrlD:
			// Ctrl+D to confirm
			if m.Message() != "" {
				m.Hide()
				return m, func() tea.Msg {
					return ConfirmMsg{
						Message: m.Message(),
						Amend:   m.amend,
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// openEditor creates a temp file and opens the external editor
func (m *Model) openEditor() tea.Cmd {
	// Create temp file with current content
	tmpPath, err := git.CreateTempCommitFile(m.textarea.Value())
	if err != nil {
		return func() tea.Msg {
			return EditorFinishedMsg{Err: err}
		}
	}
	m.tempFilePath = tmpPath

	// Get editor command
	editorCmd := git.EditorCmd(tmpPath)

	// Use tea.ExecProcess to run editor and return to TUI
	return tea.ExecProcess(editorCmd, func(err error) tea.Msg {
		content := ""
		if err == nil {
			content, _ = git.ReadTempCommitFile(tmpPath)
		}
		return EditorFinishedMsg{
			Content: strings.TrimSpace(content),
			Err:     err,
		}
	})
}

// View implements tea.Model
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	title := "Commit"
	if m.amend {
		title = "Amend Commit"
	}

	content := m.titleStyle.Render(title) + "\n\n"
	content += m.textarea.View() + "\n\n"
	content += m.helpStyle.Render("Ctrl+D to confirm • Ctrl+E to open editor • Esc to cancel")

	modal := m.borderStyle.Render(content)

	// Center the modal
	modalWidth := lipgloss.Width(modal)
	modalHeight := lipgloss.Height(modal)

	padLeft := (m.width - modalWidth) / 2
	padTop := (m.height - modalHeight) / 2

	if padLeft < 0 {
		padLeft = 0
	}
	if padTop < 0 {
		padTop = 0
	}

	// Build padding
	var b strings.Builder
	for i := 0; i < padTop; i++ {
		b.WriteString("\n")
	}
	lines := strings.Split(modal, "\n")
	for _, line := range lines {
		b.WriteString(strings.Repeat(" ", padLeft))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
