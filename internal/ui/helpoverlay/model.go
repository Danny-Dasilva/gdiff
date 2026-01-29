package helpoverlay

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha colors
var (
	colorMauve   = lipgloss.Color("#cba6f7")
	colorBase    = lipgloss.Color("#1e1e2e")
	colorBlue    = lipgloss.Color("#89b4fa")
	colorYellow  = lipgloss.Color("#f9e2af")
	colorText    = lipgloss.Color("#cdd6f4")
	colorOverlay = lipgloss.Color("#6c7086")
)

// Model represents the floating help overlay
type Model struct {
	visible bool
	width   int
	height  int
}

// New creates a new help overlay model
func New() Model {
	return Model{}
}

// Toggle toggles visibility of the help overlay
func (m *Model) Toggle() {
	m.visible = !m.visible
}

// Hide hides the help overlay
func (m *Model) Hide() {
	m.visible = false
}

// Visible returns whether the overlay is visible
func (m Model) Visible() bool {
	return m.visible
}

// SetSize sets the terminal dimensions for centering
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// View renders the help overlay. Returns empty string when hidden.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorBlue).
		Underline(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(colorText)

	dimStyle := lipgloss.NewStyle().
		Foreground(colorOverlay)

	// Build columns
	nav := m.buildSection(headerStyle, keyStyle, descStyle, "Navigation", []keybinding{
		{"j/k", "up/down"},
		{"h/l", "left/right"},
		{"gg", "top"},
		{"G", "bottom"},
		{"Ctrl+d", "half down"},
		{"Ctrl+u", "half up"},
		{"}", "next hunk"},
		{"{", "prev hunk"},
		{"Tab", "switch pane"},
	})

	staging := m.buildSection(headerStyle, keyStyle, descStyle, "Staging", []keybinding{
		{"Space", "toggle"},
		{"a", "stage file"},
		{"S", "stage hunk"},
		{"d", "discard"},
		{"v", "visual mode"},
		{"V", "visual lines"},
	})

	viewCol := m.buildSection(headerStyle, keyStyle, descStyle, "View", []keybinding{
		{"t", "staged view"},
		{"?", "help"},
		{"q", "quit"},
	})

	commitCol := m.buildSection(headerStyle, keyStyle, descStyle, "Commit", []keybinding{
		{"i", "input"},
		{"c", "dialog"},
		{"C", "amend"},
		{"Ctrl+e", "editor"},
	})

	pushCol := m.buildSection(headerStyle, keyStyle, descStyle, "Push", []keybinding{
		{"p", "push"},
		{"P", "force push"},
	})

	// Layout: two rows of columns
	colGap := "    "

	leftCols := lipgloss.JoinHorizontal(lipgloss.Top,
		nav,
		colGap,
		staging,
		colGap,
		viewCol,
	)

	rightCols := lipgloss.JoinHorizontal(lipgloss.Top,
		commitCol,
		colGap,
		pushCol,
	)

	body := lipgloss.JoinVertical(lipgloss.Left,
		leftCols,
		"",
		rightCols,
	)

	closeHint := dimStyle.Render("Press ? or Esc to close")

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText)

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("Keybindings"),
		"",
		body,
		"",
		closeHint,
	)

	// Modal box
	modalWidth := m.width * 70 / 100
	if modalWidth > 72 {
		modalWidth = 72
	}
	if modalWidth < 40 {
		modalWidth = 40
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMauve).
		Background(colorBase).
		Padding(1, 2).
		Width(modalWidth)

	modal := boxStyle.Render(content)

	// Center the modal on screen
	modalRenderedWidth := lipgloss.Width(modal)
	modalRenderedHeight := lipgloss.Height(modal)

	padLeft := (m.width - modalRenderedWidth) / 2
	padTop := (m.height - modalRenderedHeight) / 2
	if padLeft < 0 {
		padLeft = 0
	}
	if padTop < 0 {
		padTop = 0
	}

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

type keybinding struct {
	key  string
	desc string
}

func (m Model) buildSection(headerStyle, keyStyle, descStyle lipgloss.Style, title string, bindings []keybinding) string {
	var lines []string
	lines = append(lines, headerStyle.Render(title))

	for _, kb := range bindings {
		k := keyStyle.Render(lipgloss.NewStyle().Width(8).Render(kb.key))
		d := descStyle.Render(kb.desc)
		lines = append(lines, k+d)
	}

	return strings.Join(lines, "\n")
}
