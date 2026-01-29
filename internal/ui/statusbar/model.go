package statusbar

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/internal/ui/spinner"
)

// Model represents the status bar component
type Model struct {
	width       int
	mode        string
	message     string
	branch      string
	fileCount   int
	stagedCount int
	focusedPane types.Pane
	keyMap      types.KeyMap
	spinner     spinner.Model

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
		keyMap:  keyMap,
		mode:    "NORMAL",
		spinner: spinner.New(),
		// Main bar with subtle gradient effect via background
		barStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("235")),
		// Mode indicator - pill-shaped with bold colors
		modeStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("78")).
			Padding(0, 1),
		// Branch with git icon styling
		branchStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")).
			Background(lipgloss.Color("237")).
			Padding(0, 1),
		// Key hints - subtle and readable
		keyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("235")),
		// Messages - warning/info color
		messageStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Background(lipgloss.Color("235")).
			Bold(true).
			Padding(0, 1),
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

// SetFocusedPane updates which pane is focused for context-sensitive hints
func (m *Model) SetFocusedPane(pane types.Pane) {
	m.focusedPane = pane
}

// StartSpinner starts the spinner with a message
func (m *Model) StartSpinner(message string) tea.Cmd {
	return m.spinner.Start(message)
}

// StopSpinner stops the spinner
func (m *Model) StopSpinner() {
	m.spinner.Stop()
}

// IsSpinning returns whether the spinner is active
func (m *Model) IsSpinning() bool {
	return m.spinner.IsSpinning()
}

// Update handles spinner updates
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

// View renders the status bar
func (m Model) View() string {
	var parts []string

	// Mode indicator with icon (using simple Unicode)
	modeIcon := "●"
	modeColor := lipgloss.Color("78") // Green for NORMAL
	switch m.mode {
	case "STAGED":
		modeIcon = "◆"
		modeColor = lipgloss.Color("39") // Blue
	case "VISUAL":
		modeIcon = "▸"
		modeColor = lipgloss.Color("141") // Purple
	case "INSERT":
		modeIcon = "○"
		modeColor = lipgloss.Color("214") // Orange
	default:
		modeIcon = "●"
	}
	modeStyleDynamic := m.modeStyle.Background(modeColor)
	parts = append(parts, modeStyleDynamic.Render(fmt.Sprintf("%s %s", modeIcon, m.mode)))

	// Separator
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(" | ")

	// Branch with git icon (using simple Unicode)
	if m.branch != "" {
		branchIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("*")
		parts = append(parts, sep+m.branchStyle.Render(fmt.Sprintf("%s %s", branchIcon, m.branch)))
	}

	// File count badges (using simple Unicode)
	filesBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("238")).
		Padding(0, 1).
		Render(fmt.Sprintf("# %d", m.fileCount))
	parts = append(parts, sep+filesBadge)

	// Staged count badge (green if any staged)
	if m.stagedCount > 0 {
		stagedBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("78")).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("+ %d", m.stagedCount))
		parts = append(parts, " "+stagedBadge)
	}

	// Context-sensitive key hints
	keyBracket := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	keyChar := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	keyDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	renderHint := func(key, desc string) string {
		return keyBracket.Render("[") + keyChar.Render(key) + keyBracket.Render("]") + keyDesc.Render(desc+" ")
	}

	var hintStr string
	switch m.focusedPane {
	case types.PaneCommitInput:
		hintStr = renderHint("Enter", "commit") + renderHint("Esc", "cancel")
	case types.PaneDiffView:
		if m.mode == "VISUAL" {
			hintStr = renderHint("h/l", "select") + renderHint("s", "stage") + renderHint("Esc", "cancel")
		} else {
			hintStr = renderHint("Space", "stage") + renderHint("S", "hunk") + renderHint("v", "visual") + renderHint("}", "next") + renderHint("Tab", "files") + renderHint("?", "help")
		}
	default: // PaneFileTree
		hintStr = renderHint("Space", "stage") + renderHint("a", "file") + renderHint("d", "discard") + renderHint("Tab", "diff") + renderHint("i", "commit") + renderHint("?", "help")
	}
	parts = append(parts, sep+hintStr)

	left := strings.Join(parts, "")

	// Spinner or message (right-aligned)
	var right string
	if m.spinner.IsSpinning() {
		right = " " + m.spinner.View() + " "
	} else if m.message != "" {
		// Add icon based on message type (using simple Unicode)
		msgIcon := ""
		if strings.Contains(strings.ToLower(m.message), "error") {
			msgIcon = "! "
		} else if strings.Contains(strings.ToLower(m.message), "success") || strings.Contains(strings.ToLower(m.message), "committed") {
			msgIcon = "* "
		} else if strings.Contains(strings.ToLower(m.message), "warning") {
			msgIcon = "~ "
		}
		right = m.messageStyle.Render(msgIcon + m.message)
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

// HelpHeight returns the height needed for help display
func (m Model) HelpHeight() int {
	return 1
}
