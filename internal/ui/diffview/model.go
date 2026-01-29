package diffview

import (
	"fmt"
	"strings"
	"unicode/utf8"

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

	// Accessibility
	colorblind bool // Use blue/orange instead of red/green

	// Styles
	headerStyle   lipgloss.Style
	hunkStyle     lipgloss.Style
	addedStyle    lipgloss.Style
	removedStyle  lipgloss.Style
	contextStyle  lipgloss.Style
	lineNumStyle  lipgloss.Style
	selectedStyle lipgloss.Style

	// Character-level highlight colors (stored for use in render methods)
	addedHighlightBg   string
	removedHighlightBg string
	addedMarkerColor   string
	removedMarkerColor string
}

// New creates a new diff view model.
// If colorblind is true, uses blue/orange instead of red/green for accessibility.
func New(keyMap types.KeyMap, colorblind bool) Model {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Color palette: normal (red/green) vs colorblind (blue/orange)
	var (
		addedFg, addedBg       string
		removedFg, removedBg   string
		addedHiBg, removedHiBg string
	)
	if colorblind {
		// Colorblind-friendly: blue for removed, orange for added
		addedFg = "#f5a623"  // orange text
		addedBg = "#2f2a1a"  // subtle orange background
		removedFg = "#7ab4ff" // blue text
		removedBg = "#1a1f2f" // subtle blue background
		addedHiBg = "#5c4a2d" // saturated orange highlight
		removedHiBg = "#2d3a5c" // saturated blue highlight
	} else {
		// Default: Catppuccin red/green
		addedFg = "#a6e3a1"  // green text
		addedBg = "#1a2f1a"  // subtle green background
		removedFg = "#f38ba8" // red text
		removedBg = "#2f1a1a" // subtle red background
		addedHiBg = "#2d5c3a" // saturated green highlight
		removedHiBg = "#5c2d3a" // saturated red highlight
	}

	return Model{
		keyMap:     keyMap,
		viewport:   vp,
		colorblind: colorblind,
		// File header style - bold cyan with underline effect
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		// Hunk header style - blue with bold (Catppuccin blue #89b4fa)
		hunkStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true),
		// Added lines
		addedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(addedFg)).
			Background(lipgloss.Color(addedBg)),
		// Removed lines
		removedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(removedFg)).
			Background(lipgloss.Color(removedBg)),
		// Context lines - dim text (Catppuccin subtext0 #a6adc8)
		contextStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8")),
		// Line numbers - dim (Catppuccin surface2 #585b70)
		lineNumStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")),
		// Selected line highlight
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Bold(true),
		// Character-level highlight colors
		addedHighlightBg:   addedHiBg,
		removedHighlightBg: removedHiBg,
		addedMarkerColor:   addedFg,
		removedMarkerColor: removedFg,
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

		case key.Matches(msg, m.keyMap.FullPageDown):
			m.moveCursor(m.height)
			m.syncViewport()

		case key.Matches(msg, m.keyMap.FullPageUp):
			m.moveCursor(-m.height)
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

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			m.moveCursor(3)
			m.syncViewport()
		case tea.MouseButtonWheelUp:
			m.moveCursor(-3)
			m.syncViewport()
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
	// Scroll margin: keep cursor N lines from viewport edges (like vim scrolloff)
	const scrollMargin = 5

	margin := scrollMargin
	if margin > m.viewport.Height/2 {
		margin = m.viewport.Height / 2
	}

	if m.cursor < m.viewport.YOffset+margin {
		offset := m.cursor - margin
		if offset < 0 {
			offset = 0
		}
		m.viewport.SetYOffset(offset)
	} else if m.cursor >= m.viewport.YOffset+m.viewport.Height-margin {
		offset := m.cursor - m.viewport.Height + margin + 1
		total := m.totalLines()
		if offset > total-m.viewport.Height {
			offset = total - m.viewport.Height
		}
		if offset < 0 {
			offset = 0
		}
		m.viewport.SetYOffset(offset)
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

	// Side-by-side column width: split available width in half
	// Each side: line number (5) + separator (3) + content
	halfWidth := m.width / 2
	if halfWidth < 20 {
		halfWidth = 20
	}
	gutterWidth := 6  // "NNNN |"
	contentWidth := halfWidth - gutterWidth - 2 // -2 for marker and padding
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("\u2502") // vertical bar

	for _, fd := range m.diffs {
		// File header spans full width
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
			lines := hunk.Lines
			i := 0
			for i < len(lines) {
				line := lines[i]

				if line.Type == diff.LineHunkHeader {
					// Hunk header spans full width
					hunkContent := m.renderHunkHeaderSBS(line, halfWidth)
					if m.isLineSelected(lineNum) {
						hunkContent = m.selectedStyle.Render(hunkContent)
					}
					b.WriteString(hunkContent)
					b.WriteString("\n")
					lineNum++
					i++
					continue
				}

				if line.Type == diff.LineContext {
					// Context: show same line on both sides
					left := m.renderSBSSide(line.OldNum, " ", line.Content, contentWidth, m.contextStyle)
					right := m.renderSBSSide(line.NewNum, " ", line.Content, contentWidth, m.contextStyle)
					row := left + divider + right
					if m.isLineSelected(lineNum) {
						row = m.selectedStyle.Render(row)
					}
					b.WriteString(row)
					b.WriteString("\n")
					lineNum++
					i++
					continue
				}

				// Collect removed/added blocks for pairing
				var removed, added []diff.Line
				var removedStart int = lineNum
				for i < len(lines) && lines[i].Type == diff.LineRemoved {
					removed = append(removed, lines[i])
					i++
					lineNum++
				}
				addedStart := lineNum
				for i < len(lines) && lines[i].Type == diff.LineAdded {
					added = append(added, lines[i])
					i++
					lineNum++
				}

				// Compute character-level diffs for paired lines
				pairCount := min(len(removed), len(added))
				type charDiffPair struct {
					oldChanges []diff.CharChange
					newChanges []diff.CharChange
				}
				charDiffs := make([]charDiffPair, pairCount)
				for j := 0; j < pairCount; j++ {
					old, new := diff.ComputeCharDiff(removed[j].Content, added[j].Content)
					charDiffs[j] = charDiffPair{old, new}
				}

				// Render paired side-by-side
				maxLen := max(len(removed), len(added))
				for j := 0; j < maxLen; j++ {
					var left, right string
					rowLineNum := -1
					if j < len(removed) {
						var changes []diff.CharChange
						if j < pairCount {
							changes = charDiffs[j].oldChanges
						}
						left = m.renderSBSSideHighlighted(removed[j].OldNum, "-", removed[j].Content, contentWidth, m.removedStyle, changes)
						rowLineNum = removedStart + j
					} else {
						left = m.renderSBSEmpty(contentWidth)
					}
					if j < len(added) {
						var changes []diff.CharChange
						if j < pairCount {
							changes = charDiffs[j].newChanges
						}
						right = m.renderSBSSideHighlighted(added[j].NewNum, "+", added[j].Content, contentWidth, m.addedStyle, changes)
						if rowLineNum < 0 {
							rowLineNum = addedStart + j
						}
					} else {
						right = m.renderSBSEmpty(contentWidth)
						if rowLineNum < 0 {
							rowLineNum = removedStart + j
						}
					}
					row := left + divider + right
					if rowLineNum >= 0 && m.isLineSelected(rowLineNum) {
						row = m.selectedStyle.Render(row)
					}
					b.WriteString(row)
					b.WriteString("\n")
				}
			}
		}
	}

	m.viewport.SetContent(b.String())
}

// renderSBSSideHighlighted renders one side with character-level highlighting
func (m Model) renderSBSSideHighlighted(num int, marker string, content string, contentWidth int, baseStyle lipgloss.Style, changes []diff.CharChange) string {
	numStr := fmt.Sprintf("%4d", num)
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("|")

	// Truncate content to fit (rune-based to avoid splitting multibyte chars)
	runeCount := utf8.RuneCountInString(content)
	if runeCount > contentWidth {
		runes := []rune(content)
		content = string(runes[:contentWidth])
		runeCount = contentWidth
	}

	var markerStyled string
	switch marker {
	case "+":
		markerStyled = lipgloss.NewStyle().Foreground(lipgloss.Color(m.addedMarkerColor)).Bold(true).Render("+")
	case "-":
		markerStyled = lipgloss.NewStyle().Foreground(lipgloss.Color(m.removedMarkerColor)).Bold(true).Render("-")
	default:
		markerStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(" ")
	}

	// Render content with character-level highlighting
	var styledContent string
	if len(changes) == 0 {
		// No char-level changes, render uniformly
		padded := content + strings.Repeat(" ", contentWidth-runeCount)
		styledContent = baseStyle.Render(padded)
	} else {
		// Highlight changed characters with saturated background color
		var highlightStyle lipgloss.Style
		switch marker {
		case "-":
			highlightStyle = baseStyle.Background(lipgloss.Color(m.removedHighlightBg)).Bold(true)
		case "+":
			highlightStyle = baseStyle.Background(lipgloss.Color(m.addedHighlightBg)).Bold(true)
		default:
			highlightStyle = baseStyle.Bold(true)
		}
		runes := []rune(content)
		pos := 0
		var buf strings.Builder
		for _, ch := range changes {
			start := ch.Start
			end := ch.End
			if start > len(runes) {
				start = len(runes)
			}
			if end > len(runes) {
				end = len(runes)
			}
			// Render unchanged segment before this change
			if pos < start {
				buf.WriteString(baseStyle.Render(string(runes[pos:start])))
			}
			// Render changed segment with highlight
			if start < end {
				buf.WriteString(highlightStyle.Render(string(runes[start:end])))
			}
			pos = end
		}
		// Render remaining unchanged content
		if pos < len(runes) {
			buf.WriteString(baseStyle.Render(string(runes[pos:])))
		}
		// Pad to fixed width
		remaining := contentWidth - runeCount
		if remaining > 0 {
			buf.WriteString(strings.Repeat(" ", remaining))
		}
		styledContent = buf.String()
	}

	return m.lineNumStyle.Render(numStr) + " " + separator + markerStyled + styledContent
}

// renderSBSSide renders one side of a side-by-side diff line
func (m Model) renderSBSSide(num int, marker string, content string, contentWidth int, style lipgloss.Style) string {
	numStr := fmt.Sprintf("%4d", num)
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("|")

	// Truncate content to fit (rune-based to avoid splitting multibyte chars)
	runeCount := utf8.RuneCountInString(content)
	if runeCount > contentWidth {
		runes := []rune(content)
		content = string(runes[:contentWidth])
		runeCount = contentWidth
	}
	// Pad content to fixed width
	padded := content + strings.Repeat(" ", contentWidth-runeCount)

	var markerStyled string
	switch marker {
	case "+":
		markerStyled = lipgloss.NewStyle().Foreground(lipgloss.Color(m.addedMarkerColor)).Bold(true).Render("+")
	case "-":
		markerStyled = lipgloss.NewStyle().Foreground(lipgloss.Color(m.removedMarkerColor)).Bold(true).Render("-")
	default:
		markerStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(" ")
	}

	return m.lineNumStyle.Render(numStr) + " " + separator + markerStyled + style.Render(padded)
}

// renderSBSEmpty renders an empty side for unpaired lines
func (m Model) renderSBSEmpty(contentWidth int) string {
	numStr := "    "
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("|")
	padded := strings.Repeat(" ", contentWidth+1) // +1 for marker space
	return m.lineNumStyle.Render(numStr) + " " + separator + padded
}

// renderHunkHeaderSBS renders a hunk header for side-by-side view
func (m Model) renderHunkHeaderSBS(line diff.Line, halfWidth int) string {
	hunkMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true).Render("@@")
	return fmt.Sprintf("  %s %s %s", hunkMarker, m.hunkStyle.Render(line.Content), hunkMarker)
}

func (m Model) renderLine(line diff.Line, lineNum int) string {
	// Line number gutter with separator
	var numStr string
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("|")

	switch line.Type {
	case diff.LineHunkHeader:
		// Hunk header with blue bold @@ markers (Catppuccin blue #89b4fa)
		hunkMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true).Render("@@")
		return fmt.Sprintf("         %s %s %s", hunkMarker, m.hunkStyle.Render(line.Content), hunkMarker)
	case diff.LineAdded:
		numStr = fmt.Sprintf("%4s %4d", "", line.NewNum)
		addMarker := lipgloss.NewStyle().Foreground(lipgloss.Color(m.addedMarkerColor)).Bold(true).Render("+")
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + addMarker + m.addedStyle.Render(line.Content)
	case diff.LineRemoved:
		numStr = fmt.Sprintf("%4d %4s", line.OldNum, "")
		removeMarker := lipgloss.NewStyle().Foreground(lipgloss.Color(m.removedMarkerColor)).Bold(true).Render("-")
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + removeMarker + m.removedStyle.Render(line.Content)
	default:
		numStr = fmt.Sprintf("%4d %4d", line.OldNum, line.NewNum)
		contextMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(" ")
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
