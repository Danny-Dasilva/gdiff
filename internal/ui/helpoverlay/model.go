package helpoverlay

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	colorMauve   = lipgloss.Color("#cba6f7")
	colorBase    = lipgloss.Color("#1e1e2e")
	colorBlue    = lipgloss.Color("#89b4fa")
	colorYellow  = lipgloss.Color("#f9e2af")
	colorText    = lipgloss.Color("#cdd6f4")
	colorOverlay = lipgloss.Color("#6c7086")
)

type keybinding struct {
	key  string
	desc string
}

type Model struct {
	visible bool
	width   int
	height  int
}

func New() Model {
	return Model{}
}

func (m *Model) Toggle() {
	m.visible = !m.visible
}

func (m *Model) Hide() {
	m.visible = false
}

func (m Model) Visible() bool {
	return m.visible
}

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

	nav := buildSection(headerStyle, keyStyle, descStyle, "Navigation", []keybinding{
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

	staging := buildSection(headerStyle, keyStyle, descStyle, "Staging", []keybinding{
		{"Space", "toggle"},
		{"a", "stage file"},
		{"S", "stage hunk"},
		{"d", "discard"},
		{"v", "visual mode"},
		{"V", "visual lines"},
	})

	viewCol := buildSection(headerStyle, keyStyle, descStyle, "View", []keybinding{
		{"t", "staged view"},
		{"?", "help"},
		{"q", "quit"},
	})

	commitCol := buildSection(headerStyle, keyStyle, descStyle, "Commit", []keybinding{
		{"i", "input"},
		{"c", "dialog"},
		{"C", "amend"},
		{"Ctrl+e", "editor"},
	})

	pushCol := buildSection(headerStyle, keyStyle, descStyle, "Push", []keybinding{
		{"p", "push"},
		{"P", "force push"},
	})

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

	modalWidth := clamp(m.width*70/100, 40, 72)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMauve).
		Background(colorBase).
		Padding(1, 2).
		Width(modalWidth)

	modal := boxStyle.Render(content)

	padLeft := max((m.width-lipgloss.Width(modal))/2, 0)
	padTop := max((m.height-lipgloss.Height(modal))/2, 0)

	var b strings.Builder
	b.WriteString(strings.Repeat("\n", padTop))
	indent := strings.Repeat(" ", padLeft)
	for _, line := range strings.Split(modal, "\n") {
		b.WriteString(indent)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}

func buildSection(headerStyle, keyStyle, descStyle lipgloss.Style, title string, bindings []keybinding) string {
	lines := []string{headerStyle.Render(title)}

	keyWidth := lipgloss.NewStyle().Width(8)
	for _, kb := range bindings {
		k := keyStyle.Render(keyWidth.Render(kb.key))
		d := descStyle.Render(kb.desc)
		lines = append(lines, k+d)
	}

	return strings.Join(lines, "\n")
}
