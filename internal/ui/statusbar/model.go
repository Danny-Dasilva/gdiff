package statusbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/danny/gdiff/internal/types"
)

// Model represents the status bar component
type Model struct {
	width       int
	mode        string
	message     string
	branch      string
	fileCount   int
	stagedCount int
	keyMap      types.KeyMap
	showHelp    bool

	// Styles
	barStyle     lipgloss.Style
	modeStyle    lipgloss.Style
	branchStyle  lipgloss.Style
	keyStyle     lipgloss.Style
	messageStyle lipgloss.Style
}

// New creates a new status bar model
func New(keyMap types.KeyMap) Model {
	return Model{
		keyMap:       keyMap,
		mode:         "NORMAL",
		barStyle:     lipgloss.NewStyle().Background(lipgloss.Color("236")),
		modeStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("40")).Background(lipgloss.Color("236")),
		branchStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Background(lipgloss.Color("236")),
		keyStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Background(lipgloss.Color("236")),
		messageStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Background(lipgloss.Color("236")),
	}
}

// SetWidth updates the width
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetBranch updates the branch name
func (m *Model) SetBranch(branch string) {
	m.branch = branch
}

// SetMessage sets a temporary message
func (m *Model) SetMessage(msg string) {
	m.message = msg
}

// ClearMessage clears the message
func (m *Model) ClearMessage() {
	m.message = ""
}

// SetCounts updates file counts
func (m *Model) SetCounts(total, staged int) {
	m.fileCount = total
	m.stagedCount = staged
}

// SetMode updates the current mode
func (m *Model) SetMode(mode string) {
	m.mode = mode
}

// ToggleHelp toggles the help display
func (m *Model) ToggleHelp() {
	m.showHelp = !m.showHelp
}

// View renders the status bar
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	var parts []string

	// Mode indicator
	parts = append(parts, m.modeStyle.Render(fmt.Sprintf(" %s ", m.mode)))

	// Branch
	if m.branch != "" {
		parts = append(parts, m.branchStyle.Render(fmt.Sprintf(" \ue0a0 %s ", m.branch)))
	}

	// File counts
	countStr := fmt.Sprintf(" %d files", m.fileCount)
	if m.stagedCount > 0 {
		countStr = fmt.Sprintf(" %d/%d staged", m.stagedCount, m.fileCount)
	}
	parts = append(parts, m.keyStyle.Render(countStr))

	// Key hints
	hints := " | [a]stage [A]unstage [c]ommit [p]ush | j/k nav | ? help"
	parts = append(parts, m.keyStyle.Render(hints))

	left := strings.Join(parts, "")

	// Message (right-aligned)
	var right string
	if m.message != "" {
		right = m.messageStyle.Render(" " + m.message + " ")
	}

	// Calculate padding
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	padding := m.width - leftLen - rightLen
	if padding < 0 {
		padding = 0
	}

	return m.barStyle.Width(m.width).Render(left + strings.Repeat(" ", padding) + right)
}

func (m Model) renderHelp() string {
	help := `
 Navigation                     Staging                         Actions
 ──────────                     ───────                         ───────
 j/k      up/down               a        stage file             c   commit
 h/l      left/right            A        unstage file           C   amend
 gg/G     top/bottom            s/space  stage selection        p   push
 ^d/^u    half page             u        unstage selection      P   force push
 }/{      next/prev hunk        x        revert (confirm)       q   quit
 [c/]c    next/prev change      v        visual mode            ?   close help
 tab      switch pane           V        visual line
`
	return m.barStyle.Width(m.width).Render(help)
}

// HelpHeight returns the height needed for help display
func (m Model) HelpHeight() int {
	if m.showHelp {
		return 12
	}
	return 1
}
