package diffview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

// Model represents the diff view component
type Model struct {
	diffs    []diff.FileDiff
	path     string
	viewport viewport.Model
	width    int
	height   int
	focused  bool
	keyMap   types.KeyMap

	// Cursor position within diff
	cursor     int
	hunkIndex  int
	lineIndex  int

	// Visual mode selection
	visualMode  bool
	selectStart int
	selectEnd   int

	// Character-level selection (within a line)
	charMode   bool // True when in character selection mode
	charStart  int  // Start character position in current line
	charCursor int  // Current character cursor position

	// Styles
	headerStyle   lipgloss.Style
	hunkStyle     lipgloss.Style
	addedStyle    lipgloss.Style
	removedStyle  lipgloss.Style
	contextStyle  lipgloss.Style
	lineNumStyle  lipgloss.Style
	selectedStyle lipgloss.Style
}

// New creates a new diff view model
func New(keyMap types.KeyMap) Model {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	return Model{
		keyMap:   keyMap,
		viewport: vp,
		// File header style - bold cyan with underline effect
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		// Hunk header style - purple/magenta with italic
		hunkStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")).
			Italic(true).
			Background(lipgloss.Color("235")),
		// Added lines - green foreground with subtle green background
		addedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Background(lipgloss.Color("22")),
		// Removed lines - red foreground with subtle red background
		removedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("210")).
			Background(lipgloss.Color("52")),
		// Context lines - dimmed
		contextStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		// Line numbers - dimmed and right-aligned feel
		lineNumStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("235")),
		// Selected line highlight
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Bold(true),
	}
}

// SetDiff updates the displayed diff
func (m *Model) SetDiff(path string, diffs []diff.FileDiff) {
	m.path = path
	m.diffs = diffs
	m.cursor = 0
	m.hunkIndex = 0
	m.lineIndex = 0
	m.visualMode = false
	m.updateViewportContent()
}

// SetSize updates the component dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	m.updateViewportContent()
}

// SetFocused updates the focus state
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Down):
			m.moveCursor(1)
			m.syncViewport()

		case key.Matches(msg, m.keyMap.Up):
			m.moveCursor(-1)
			m.syncViewport()

		case key.Matches(msg, m.keyMap.HalfDown):
			m.moveCursor(m.height / 2)
			m.syncViewport()

		case key.Matches(msg, m.keyMap.HalfUp):
			m.moveCursor(-m.height / 2)
			m.syncViewport()

		case key.Matches(msg, m.keyMap.Top):
			m.cursor = 0
			m.syncViewport()

		case key.Matches(msg, m.keyMap.Bottom):
			m.cursor = m.totalLines() - 1
			m.syncViewport()

		case key.Matches(msg, m.keyMap.NextHunk):
			m.nextHunk()
			m.syncViewport()

		case key.Matches(msg, m.keyMap.PrevHunk):
			m.prevHunk()
			m.syncViewport()

		case key.Matches(msg, m.keyMap.VisualMode):
			m.toggleVisualMode()

		case key.Matches(msg, m.keyMap.Escape):
			// Exit visual/character mode
			if m.charMode || m.visualMode {
				m.charMode = false
				m.visualMode = false
				m.charStart = 0
				m.charCursor = 0
			}

		case key.Matches(msg, m.keyMap.Right):
			// In character mode, move character cursor right
			if m.charMode {
				m.moveCharCursor(1)
			}

		case key.Matches(msg, m.keyMap.Left):
			// In character mode, move character cursor left
			if m.charMode {
				m.moveCharCursor(-1)
			}

		default:
			m.viewport, cmd = m.viewport.Update(msg)
		}
	}

	return m, cmd
}

func (m *Model) moveCursor(delta int) {
	total := m.totalLines()
	if total == 0 {
		return
	}

	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= total {
		m.cursor = total - 1
	}

	if m.visualMode {
		m.selectEnd = m.cursor
	}
}

func (m *Model) syncViewport() {
	// Ensure cursor is visible in viewport
	if m.cursor < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor)
	} else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(m.cursor - m.viewport.Height + 1)
	}
	m.updateViewportContent()
}

func (m *Model) totalLines() int {
	count := 0
	for _, fd := range m.diffs {
		count++ // File header
		for _, hunk := range fd.Hunks {
			count += len(hunk.Lines)
		}
	}
	return count
}

func (m *Model) nextHunk() {
	lineNum := 0
	for _, fd := range m.diffs {
		lineNum++ // File header
		for _, hunk := range fd.Hunks {
			// Find first line of next hunk after cursor
			if lineNum > m.cursor {
				m.cursor = lineNum
				return
			}
			lineNum += len(hunk.Lines)
		}
	}
}

func (m *Model) prevHunk() {
	lineNum := 0
	lastHunkStart := 0
	for _, fd := range m.diffs {
		lineNum++ // File header
		for _, hunk := range fd.Hunks {
			if lineNum >= m.cursor {
				m.cursor = lastHunkStart
				return
			}
			lastHunkStart = lineNum
			lineNum += len(hunk.Lines)
		}
	}
}

func (m *Model) toggleVisualMode() {
	m.visualMode = !m.visualMode
	if m.visualMode {
		m.selectStart = m.cursor
		m.selectEnd = m.cursor

		// Enable character mode if on a changeable line (added/removed)
		if m.isCurrentLineChangeable() {
			m.charMode = true
			m.charStart = 0
			m.charCursor = 0
		}
	} else {
		// Exiting visual mode also exits character mode
		m.charMode = false
		m.charStart = 0
		m.charCursor = 0
	}
}

func (m *Model) updateViewportContent() {
	if len(m.diffs) == 0 {
		m.viewport.SetContent(m.contextStyle.Render("No diff to display"))
		return
	}

	var b strings.Builder
	lineNum := 0

	for _, fd := range m.diffs {
		// File header
		header := fmt.Sprintf("--- %s\n+++ %s", fd.OldPath, fd.NewPath)
		if m.isLineSelected(lineNum) {
			b.WriteString(m.selectedStyle.Render(m.headerStyle.Render(header)))
		} else {
			b.WriteString(m.headerStyle.Render(header))
		}
		b.WriteString("\n")
		lineNum++

		if fd.IsBinary {
			b.WriteString(m.contextStyle.Render("Binary file differs"))
			b.WriteString("\n")
			continue
		}

		for _, hunk := range fd.Hunks {
			for _, line := range hunk.Lines {
				lineContent := m.renderLine(line, lineNum)
				if m.isLineSelected(lineNum) {
					lineContent = m.selectedStyle.Render(lineContent)
				}
				b.WriteString(lineContent)
				b.WriteString("\n")
				lineNum++
			}
		}
	}

	m.viewport.SetContent(b.String())
}

func (m Model) renderLine(line diff.Line, lineNum int) string {
	// Line number gutter with separator
	var numStr string
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("|")

	switch line.Type {
	case diff.LineHunkHeader:
		// Hunk header with decorative markers
		hunkMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true).Render("@@")
		return fmt.Sprintf("         %s %s %s", hunkMarker, m.hunkStyle.Render(line.Content), hunkMarker)
	case diff.LineAdded:
		numStr = fmt.Sprintf("%4s %4d", "", line.NewNum)
		addMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true).Render("+")
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + addMarker + m.addedStyle.Render(line.Content)
	case diff.LineRemoved:
		numStr = fmt.Sprintf("%4d %4s", line.OldNum, "")
		removeMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Bold(true).Render("-")
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + removeMarker + m.removedStyle.Render(line.Content)
	default:
		numStr = fmt.Sprintf("%4d %4d", line.OldNum, line.NewNum)
		contextMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" ")
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + contextMarker + m.contextStyle.Render(line.Content)
	}
}

func (m Model) isLineSelected(lineNum int) bool {
	if lineNum == m.cursor && m.focused {
		return true
	}
	if m.visualMode {
		start, end := m.selectStart, m.selectEnd
		if start > end {
			start, end = end, start
		}
		return lineNum >= start && lineNum <= end
	}
	return false
}

// View implements tea.Model
func (m Model) View() string {
	return m.viewport.View()
}

// CharSelection represents a character-level selection within a line
type CharSelection struct {
	LineIndex int // Index in flattened line list
	Start     int // Start character position
	End       int // End character position
}

// GetCharacterSelection returns the current character selection if in char mode
func (m Model) GetCharacterSelection() *CharSelection {
	if !m.charMode {
		return nil
	}

	start, end := m.charStart, m.charCursor
	if start > end {
		start, end = end, start
	}

	return &CharSelection{
		LineIndex: m.cursor,
		Start:     start,
		End:       end,
	}
}

// getCurrentLineContent returns the content of the line at the current cursor position
func (m Model) getCurrentLineContent() (string, diff.LineType) {
	lineNum := 0
	for _, fd := range m.diffs {
		if lineNum == m.cursor {
			return "", diff.LineContext // File header
		}
		lineNum++

		for _, hunk := range fd.Hunks {
			for _, line := range hunk.Lines {
				if lineNum == m.cursor {
					return line.Content, line.Type
				}
				lineNum++
			}
		}
	}
	return "", diff.LineContext
}

// isCurrentLineChangeable returns true if the current line is an added or removed line
func (m Model) isCurrentLineChangeable() bool {
	_, lineType := m.getCurrentLineContent()
	return lineType == diff.LineAdded || lineType == diff.LineRemoved
}

// moveCharCursor moves the character cursor by delta positions
func (m *Model) moveCharCursor(delta int) {
	content, _ := m.getCurrentLineContent()
	maxPos := len([]rune(content))

	m.charCursor += delta
	if m.charCursor < 0 {
		m.charCursor = 0
	}
	if m.charCursor > maxPos {
		m.charCursor = maxPos
	}
}

// CharStagingInfo contains all info needed to stage characters
type CharStagingInfo struct {
	Hunk          diff.Hunk
	HunkLineIndex int // Index of line within hunk
	CharStart     int
	CharEnd       int
}

// GetCharStagingInfo returns the info needed to stage the current character selection
// Returns nil if not in character mode or selection is invalid
func (m Model) GetCharStagingInfo() *CharStagingInfo {
	if !m.charMode {
		return nil
	}

	// Find the hunk and line index for current cursor position
	lineNum := 0
	for _, fd := range m.diffs {
		lineNum++ // File header

		for _, hunk := range fd.Hunks {
			hunkStartLine := lineNum
			for i, line := range hunk.Lines {
				if lineNum == m.cursor {
					// Found the line - check if it's a changeable line
					if line.Type != diff.LineAdded && line.Type != diff.LineRemoved {
						return nil
					}

					start, end := m.charStart, m.charCursor
					if start > end {
						start, end = end, start
					}

					return &CharStagingInfo{
						Hunk:          hunk,
						HunkLineIndex: i,
						CharStart:     start,
						CharEnd:       end,
					}
				}
				lineNum++
			}
			_ = hunkStartLine // silence unused warning
		}
	}

	return nil
}

// IsInCharMode returns true if currently in character selection mode
func (m Model) IsInCharMode() bool {
	return m.charMode
}
