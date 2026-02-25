package statusbar

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/internal/ui/spinner"
)

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

	barStyle     lipgloss.Style
	modeStyle    lipgloss.Style
	branchStyle  lipgloss.Style
	keyStyle     lipgloss.Style
	messageStyle lipgloss.Style
}

func New(keyMap types.KeyMap) Model {
	return Model{
		keyMap:  keyMap,
		mode:    "NORMAL",
		spinner: spinner.New(),
		barStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("235")),
		modeStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("78")).
			Padding(0, 1),
		branchStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")).
			Background(lipgloss.Color("237")).
			Padding(0, 1),
		keyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("235")),
		messageStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Background(lipgloss.Color("235")).
			Bold(true).
			Padding(0, 1),
	}
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) SetBranch(branch string) {
	m.branch = branch
}

func (m *Model) SetMessage(msg string) {
	m.message = msg
}

func (m *Model) ClearMessage() {
	m.message = ""
}

func (m *Model) SetCounts(total, staged int) {
	m.fileCount = total
	m.stagedCount = staged
}

func (m *Model) SetMode(mode string) {
	m.mode = mode
}

func (m *Model) SetFocusedPane(pane types.Pane) {
	m.focusedPane = pane
}

func (m *Model) StartSpinner(message string) tea.Cmd {
	return m.spinner.Start(message)
}

func (m *Model) StopSpinner() {
	m.spinner.Stop()
}

func (m *Model) IsSpinning() bool {
	return m.spinner.IsSpinning()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

func (m Model) View() string {
	var parts []string

	modeIcon := "●"
	modeColor := lipgloss.Color("78")
	switch m.mode {
	case "STAGED":
		modeIcon = "◆"
		modeColor = lipgloss.Color("39")
	case "VISUAL":
		modeIcon = "▸"
		modeColor = lipgloss.Color("141")
	case "INSERT":
		modeIcon = "○"
		modeColor = lipgloss.Color("214")
	}
	modeStyleDynamic := m.modeStyle.Background(modeColor)
	parts = append(parts, modeStyleDynamic.Render(fmt.Sprintf("%s %s", modeIcon, m.mode)))

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(" | ")

	if m.branch != "" {
		branchIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("*")
		parts = append(parts, sep+m.branchStyle.Render(fmt.Sprintf("%s %s", branchIcon, m.branch)))
	}

	filesBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("238")).
		Padding(0, 1).
		Render(fmt.Sprintf("# %d", m.fileCount))
	parts = append(parts, sep+filesBadge)

	if m.stagedCount > 0 {
		stagedBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("78")).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("+ %d", m.stagedCount))
		parts = append(parts, " "+stagedBadge)
	}

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
	default:
		hintStr = renderHint("Space", "stage") + renderHint("a", "file") + renderHint("d", "discard") + renderHint("Tab", "diff") + renderHint("i", "commit") + renderHint("?", "help")
	}
	parts = append(parts, sep+hintStr)

	left := strings.Join(parts, "")

	var right string
	if m.spinner.IsSpinning() {
		right = " " + m.spinner.View() + " "
	} else if m.message != "" {
		msgLower := strings.ToLower(m.message)
		msgIcon := ""
		switch {
		case strings.Contains(msgLower, "error"):
			msgIcon = "! "
		case strings.Contains(msgLower, "success"), strings.Contains(msgLower, "committed"):
			msgIcon = "* "
		case strings.Contains(msgLower, "warning"):
			msgIcon = "~ "
		}
		right = m.messageStyle.Render(msgIcon + m.message)
	}

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	padding := m.width - leftLen - rightLen
	if padding < 0 {
		padding = 0
	}

	return m.barStyle.Width(m.width).Render(left + strings.Repeat(" ", padding) + right)
}

func (m Model) HelpHeight() int {
	return 1
}
