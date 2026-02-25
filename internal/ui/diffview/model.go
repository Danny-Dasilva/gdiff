package diffview

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

type Model struct {
	diffs    []diff.FileDiff
	path     string
	viewport viewport.Model
	width    int
	height   int
	focused  bool
	keyMap   types.KeyMap

	cursor     int
	hunkIndex  int
	lineIndex  int

	visualMode  bool
	selectStart int
	selectEnd   int

	charMode   bool
	charStart  int
	charCursor int

	colorblind bool

	headerStyle    lipgloss.Style
	hunkStyle      lipgloss.Style
	addedStyle     lipgloss.Style
	removedStyle   lipgloss.Style
	contextStyle   lipgloss.Style
	lineNumStyle   lipgloss.Style
	selectedStyle  lipgloss.Style
	separatorStyle lipgloss.Style

	addedHighlightBg   string
	removedHighlightBg string
	addedMarkerColor   string
	removedMarkerColor string
}

func New(keyMap types.KeyMap, colorblind bool) Model {
	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	vp.SetContent("")

	var (
		addedFg, addedBg       string
		removedFg, removedBg   string
		addedHiBg, removedHiBg string
	)
	if colorblind {
		addedFg = "#f5a623"   // orange
		addedBg = "#2f2a1a"
		removedFg = "#7ab4ff" // blue
		removedBg = "#1a1f2f"
		addedHiBg = "#5c4a2d"
		removedHiBg = "#2d3a5c"
	} else {
		addedFg = "#a6e3a1"   // green
		addedBg = "#1a2f1a"
		removedFg = "#f38ba8" // red
		removedBg = "#2f1a1a"
		addedHiBg = "#2d5c3a"
		removedHiBg = "#5c2d3a"
	}

	return Model{
		keyMap:     keyMap,
		viewport:   vp,
		colorblind: colorblind,
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		hunkStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true),
		addedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(addedFg)).
			Background(lipgloss.Color(addedBg)),
		removedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(removedFg)).
			Background(lipgloss.Color(removedBg)),
		contextStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8")),
		lineNumStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585b70")),
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Bold(true),
		addedHighlightBg:   addedHiBg,
		removedHighlightBg: removedHiBg,
		addedMarkerColor:   addedFg,
		removedMarkerColor: removedFg,
		separatorStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
	}
}

func (m *Model) SetDiff(path string, diffs []diff.FileDiff) {
	m.path = path
	m.diffs = diffs
	m.cursor = 0
	m.hunkIndex = 0
	m.lineIndex = 0
	m.visualMode = false
	m.updateViewportContent()
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height)
	m.updateViewportContent()
}

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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
			if m.charMode || m.visualMode {
				m.charMode = false
				m.visualMode = false
				m.charStart = 0
				m.charCursor = 0
			}

		case key.Matches(msg, m.keyMap.Right):
			if m.charMode {
				m.moveCharCursor(1)
			}

		case key.Matches(msg, m.keyMap.Left):
			if m.charMode {
				m.moveCharCursor(-1)
			}

		default:
			m.viewport, cmd = m.viewport.Update(msg)
		}

	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelDown:
			m.moveCursor(3)
			m.syncViewport()
		case tea.MouseWheelUp:
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
	const scrollMargin = 5

	margin := scrollMargin
	if margin > m.viewport.Height()/2 {
		margin = m.viewport.Height() / 2
	}

	if m.cursor < m.viewport.YOffset()+margin {
		offset := m.cursor - margin
		if offset < 0 {
			offset = 0
		}
		m.viewport.SetYOffset(offset)
	} else if m.cursor >= m.viewport.YOffset()+m.viewport.Height()-margin {
		offset := m.cursor - m.viewport.Height() + margin + 1
		total := m.totalLines()
		if offset > total-m.viewport.Height() {
			offset = total - m.viewport.Height()
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

		if m.isCurrentLineChangeable() {
			m.charMode = true
			m.charStart = 0
			m.charCursor = 0
		}
	} else {
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

	halfWidth := m.width / 2
	if halfWidth < 20 {
		halfWidth = 20
	}
	gutterWidth := 6
	contentWidth := halfWidth - gutterWidth - 2
	divider := m.separatorStyle.Render("\u2502")

	for _, fd := range m.diffs {
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

				var removed, added []diff.Line
				removedStart := lineNum
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

func (m Model) renderMarker(marker string) string {
	switch marker {
	case "+":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.addedMarkerColor)).Bold(true).Render("+")
	case "-":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(m.removedMarkerColor)).Bold(true).Render("-")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")).Render(" ")
	}
}

func (m Model) renderSBSSideHighlighted(num int, marker string, content string, contentWidth int, baseStyle lipgloss.Style, changes []diff.CharChange) string {
	numStr := fmt.Sprintf("%4d", num)
	separator := m.separatorStyle.Render("|")

	runeCount := utf8.RuneCountInString(content)
	if runeCount > contentWidth {
		runes := []rune(content)
		content = string(runes[:contentWidth])
		runeCount = contentWidth
	}

	var styledContent string
	if len(changes) == 0 {
		padded := content + strings.Repeat(" ", contentWidth-runeCount)
		styledContent = baseStyle.Render(padded)
	} else {
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
			start := min(ch.Start, len(runes))
			end := min(ch.End, len(runes))
			if pos < start {
				buf.WriteString(baseStyle.Render(string(runes[pos:start])))
			}
			if start < end {
				buf.WriteString(highlightStyle.Render(string(runes[start:end])))
			}
			pos = end
		}
		if pos < len(runes) {
			buf.WriteString(baseStyle.Render(string(runes[pos:])))
		}
		if remaining := contentWidth - runeCount; remaining > 0 {
			buf.WriteString(strings.Repeat(" ", remaining))
		}
		styledContent = buf.String()
	}

	return m.lineNumStyle.Render(numStr) + " " + separator + m.renderMarker(marker) + styledContent
}

func (m Model) renderSBSSide(num int, marker string, content string, contentWidth int, style lipgloss.Style) string {
	return m.renderSBSSideHighlighted(num, marker, content, contentWidth, style, nil)
}

func (m Model) renderSBSEmpty(contentWidth int) string {
	separator := m.separatorStyle.Render("|")
	padded := strings.Repeat(" ", contentWidth+1) // +1 for marker space
	return m.lineNumStyle.Render("    ") + " " + separator + padded
}

func (m Model) renderHunkHeaderSBS(line diff.Line, halfWidth int) string {
	hunkMarker := m.hunkStyle.Render("@@")
	return fmt.Sprintf("  %s %s %s", hunkMarker, m.hunkStyle.Render(line.Content), hunkMarker)
}

func (m Model) renderLine(line diff.Line, lineNum int) string {
	separator := m.separatorStyle.Render("|")

	switch line.Type {
	case diff.LineHunkHeader:
		hunkMarker := m.hunkStyle.Render("@@")
		return fmt.Sprintf("         %s %s %s", hunkMarker, m.hunkStyle.Render(line.Content), hunkMarker)
	case diff.LineAdded:
		numStr := fmt.Sprintf("%4s %4d", "", line.NewNum)
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + m.renderMarker("+") + m.addedStyle.Render(line.Content)
	case diff.LineRemoved:
		numStr := fmt.Sprintf("%4d %4s", line.OldNum, "")
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + m.renderMarker("-") + m.removedStyle.Render(line.Content)
	default:
		numStr := fmt.Sprintf("%4d %4d", line.OldNum, line.NewNum)
		return m.lineNumStyle.Render(numStr) + " " + separator + " " + m.renderMarker(" ") + m.contextStyle.Render(line.Content)
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

func (m Model) View() string {
	return m.viewport.View()
}

type CharSelection struct {
	LineIndex int // Index in flattened line list
	Start     int // Start character position
	End       int // End character position
}

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

func (m Model) isCurrentLineChangeable() bool {
	_, lineType := m.getCurrentLineContent()
	return lineType == diff.LineAdded || lineType == diff.LineRemoved
}

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

func (m Model) GetCharStagingInfo() *CharStagingInfo {
	if !m.charMode {
		return nil
	}

	lineNum := 0
	for _, fd := range m.diffs {
		lineNum++

		for _, hunk := range fd.Hunks {
			for i, line := range hunk.Lines {
				if lineNum == m.cursor {
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
		}
	}

	return nil
}

func (m Model) IsInCharMode() bool {
	return m.charMode
}
