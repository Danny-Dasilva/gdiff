package filetree

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/danny/gdiff/internal/types"
	"github.com/danny/gdiff/pkg/diff"
)

// Model represents the file tree component
type Model struct {
	files    []diff.FileEntry
	cursor   int
	width    int
	height   int
	focused  bool
	keyMap   types.KeyMap

	// Styles
	normalStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	focusedStyle  lipgloss.Style
	statusStyles  map[diff.FileStatus]lipgloss.Style
}

// New creates a new file tree model
func New(keyMap types.KeyMap) Model {
	return Model{
		keyMap:        keyMap,
		normalStyle:   lipgloss.NewStyle(),
		selectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("238")),
		focusedStyle:  lipgloss.NewStyle().Background(lipgloss.Color("62")),
		statusStyles: map[diff.FileStatus]lipgloss.Style{
			diff.StatusModified:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
			diff.StatusAdded:     lipgloss.NewStyle().Foreground(lipgloss.Color("40")),
			diff.StatusDeleted:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
			diff.StatusRenamed:   lipgloss.NewStyle().Foreground(lipgloss.Color("141")),
			diff.StatusUntracked: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
			diff.StatusUnmerged:  lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
		},
	}
}

// SetFiles updates the file list
func (m *Model) SetFiles(files []diff.FileEntry) {
	m.files = files
	if m.cursor >= len(files) {
		m.cursor = max(0, len(files)-1)
	}
}

// SetSize updates the component dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetFocused updates the focus state
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// SelectedFile returns the currently selected file
func (m Model) SelectedFile() *diff.FileEntry {
	if len(m.files) == 0 || m.cursor < 0 || m.cursor >= len(m.files) {
		return nil
	}
	return &m.files[m.cursor]
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Down):
			if m.cursor < len(m.files)-1 {
				m.cursor++
				return m, m.emitFileSelected()
			}

		case key.Matches(msg, m.keyMap.Up):
			if m.cursor > 0 {
				m.cursor--
				return m, m.emitFileSelected()
			}

		case key.Matches(msg, m.keyMap.Top):
			m.cursor = 0
			return m, m.emitFileSelected()

		case key.Matches(msg, m.keyMap.Bottom):
			if len(m.files) > 0 {
				m.cursor = len(m.files) - 1
				return m, m.emitFileSelected()
			}

		case key.Matches(msg, m.keyMap.HalfDown):
			m.cursor = min(m.cursor+m.height/2, len(m.files)-1)
			return m, m.emitFileSelected()

		case key.Matches(msg, m.keyMap.HalfUp):
			m.cursor = max(m.cursor-m.height/2, 0)
			return m, m.emitFileSelected()
		}
	}

	return m, nil
}

func (m Model) emitFileSelected() tea.Cmd {
	if f := m.SelectedFile(); f != nil {
		return func() tea.Msg {
			return types.FileSelectedMsg{
				Path:   f.Path,
				Staged: f.Staged,
			}
		}
	}
	return nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var b strings.Builder

	// Calculate viewport
	start := 0
	if m.cursor >= m.height {
		start = m.cursor - m.height + 1
	}
	end := min(start+m.height, len(m.files))

	for i := start; i < end; i++ {
		file := m.files[i]

		// Build the line
		statusStyle := m.statusStyles[file.Status]
		if statusStyle.Value() == "" {
			statusStyle = m.normalStyle
		}

		// Status indicator
		var indicators string
		if file.Staged {
			indicators = fmt.Sprintf("%s ", file.IndexStatus.String())
		} else {
			indicators = fmt.Sprintf(" %s", file.WorkStatus.String())
		}

		// Truncate path if needed
		path := file.Path
		maxPathLen := m.width - 4 // Account for indicators and padding
		if len(path) > maxPathLen && maxPathLen > 3 {
			path = "..." + path[len(path)-maxPathLen+3:]
		}

		line := fmt.Sprintf(" %s %s", statusStyle.Render(indicators), path)

		// Apply selection style
		if i == m.cursor {
			if m.focused {
				line = m.focusedStyle.Width(m.width).Render(line)
			} else {
				line = m.selectedStyle.Width(m.width).Render(line)
			}
		} else {
			line = m.normalStyle.Width(m.width).Render(line)
		}

		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// Pad remaining lines if needed
	rendered := strings.Count(b.String(), "\n") + 1
	for i := rendered; i < m.height; i++ {
		b.WriteString("\n")
		b.WriteString(m.normalStyle.Width(m.width).Render(""))
	}

	return b.String()
}
