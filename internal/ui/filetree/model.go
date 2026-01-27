package filetree

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

// Row types for the virtual list
type rowType int

const (
	rowStagedHeader rowType = iota
	rowStagedFile
	rowChangesHeader
	rowChangesFile
)

type listRow struct {
	rowType   rowType
	fileIndex int // Index into staged or unstaged slice
}

// Model represents the file tree component
type Model struct {
	files    []diff.FileEntry
	staged   []diff.FileEntry
	unstaged []diff.FileEntry
	rows     []listRow // Virtual list of rows

	cursor  int
	width   int
	height  int
	focused bool
	keyMap  types.KeyMap

	// Section collapse state
	stagedCollapsed   bool
	unstagedCollapsed bool

	// Styles
	normalStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	focusedStyle  lipgloss.Style
	headerStyle   lipgloss.Style
	countStyle    lipgloss.Style
	iconStyle     lipgloss.Style
	statusStyles  map[diff.FileStatus]lipgloss.Style
}

// New creates a new file tree model
func New(keyMap types.KeyMap) Model {
	return Model{
		keyMap:        keyMap,
		normalStyle:   lipgloss.NewStyle(),
		selectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("238")),
		focusedStyle:  lipgloss.NewStyle().Background(lipgloss.Color("62")),
		headerStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")),
		countStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		iconStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
		statusStyles: map[diff.FileStatus]lipgloss.Style{
			diff.StatusModified:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // Orange
			diff.StatusAdded:     lipgloss.NewStyle().Foreground(lipgloss.Color("78")),  // Green
			diff.StatusDeleted:   lipgloss.NewStyle().Foreground(lipgloss.Color("204")), // Red
			diff.StatusRenamed:   lipgloss.NewStyle().Foreground(lipgloss.Color("141")), // Purple
			diff.StatusUntracked: lipgloss.NewStyle().Foreground(lipgloss.Color("78")),  // Green (new file)
			diff.StatusUnmerged:  lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
		},
	}
}

// SetFiles updates the file list and rebuilds sections
func (m *Model) SetFiles(files []diff.FileEntry) {
	m.files = files
	m.staged = nil
	m.unstaged = nil

	// Separate staged and unstaged
	for _, f := range files {
		if f.Staged {
			m.staged = append(m.staged, f)
		} else {
			m.unstaged = append(m.unstaged, f)
		}
	}

	m.rebuildRows()

	if m.cursor >= len(m.rows) {
		m.cursor = max(0, len(m.rows)-1)
	}
}

// rebuildRows creates the virtual list based on collapse state
func (m *Model) rebuildRows() {
	m.rows = nil

	// Staged section
	if len(m.staged) > 0 {
		m.rows = append(m.rows, listRow{rowType: rowStagedHeader})
		if !m.stagedCollapsed {
			for i := range m.staged {
				m.rows = append(m.rows, listRow{rowType: rowStagedFile, fileIndex: i})
			}
		}
	}

	// Changes section
	if len(m.unstaged) > 0 {
		m.rows = append(m.rows, listRow{rowType: rowChangesHeader})
		if !m.unstagedCollapsed {
			for i := range m.unstaged {
				m.rows = append(m.rows, listRow{rowType: rowChangesFile, fileIndex: i})
			}
		}
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
	if len(m.rows) == 0 || m.cursor < 0 || m.cursor >= len(m.rows) {
		return nil
	}

	row := m.rows[m.cursor]
	switch row.rowType {
	case rowStagedFile:
		if row.fileIndex < len(m.staged) {
			return &m.staged[row.fileIndex]
		}
	case rowChangesFile:
		if row.fileIndex < len(m.unstaged) {
			return &m.unstaged[row.fileIndex]
		}
	}
	return nil
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
			if m.cursor < len(m.rows)-1 {
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
			if len(m.rows) > 0 {
				m.cursor = len(m.rows) - 1
				return m, m.emitFileSelected()
			}

		case key.Matches(msg, m.keyMap.HalfDown):
			m.cursor = min(m.cursor+m.height/2, len(m.rows)-1)
			return m, m.emitFileSelected()

		case key.Matches(msg, m.keyMap.HalfUp):
			m.cursor = max(m.cursor-m.height/2, 0)
			return m, m.emitFileSelected()

		// Toggle collapse on Enter or Space when on header
		case msg.String() == "enter" || msg.String() == " ":
			if len(m.rows) > 0 && m.cursor < len(m.rows) {
				row := m.rows[m.cursor]
				if row.rowType == rowStagedHeader {
					m.stagedCollapsed = !m.stagedCollapsed
					m.rebuildRows()
				} else if row.rowType == rowChangesHeader {
					m.unstagedCollapsed = !m.unstagedCollapsed
					m.rebuildRows()
				}
			}
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

// getFileIcon returns an icon based on file extension
func getFileIcon(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "󰟓"
	case ".js", ".jsx":
		return ""
	case ".ts", ".tsx":
		return ""
	case ".py":
		return ""
	case ".rs":
		return ""
	case ".rb":
		return ""
	case ".java":
		return ""
	case ".c", ".h":
		return ""
	case ".cpp", ".hpp", ".cc":
		return ""
	case ".json":
		return ""
	case ".yaml", ".yml":
		return ""
	case ".toml":
		return ""
	case ".md":
		return ""
	case ".html":
		return ""
	case ".css", ".scss", ".sass":
		return ""
	case ".sh", ".bash", ".zsh":
		return ""
	case ".sql":
		return ""
	case ".docker", "dockerfile":
		return ""
	case ".git", ".gitignore":
		return ""
	case ".mod", ".sum":
		return "󰟓"
	default:
		return ""
	}
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
	end := min(start+m.height, len(m.rows))

	for i := start; i < end; i++ {
		row := m.rows[i]
		var line string

		switch row.rowType {
		case rowStagedHeader:
			chevron := "▼"
			if m.stagedCollapsed {
				chevron = "▶"
			}
			header := fmt.Sprintf(" %s STAGED CHANGES", chevron)
			count := m.countStyle.Render(fmt.Sprintf(" (%d)", len(m.staged)))
			line = m.headerStyle.Render(header) + count

		case rowChangesHeader:
			chevron := "▼"
			if m.unstagedCollapsed {
				chevron = "▶"
			}
			header := fmt.Sprintf(" %s CHANGES", chevron)
			count := m.countStyle.Render(fmt.Sprintf(" (%d)", len(m.unstaged)))
			line = m.headerStyle.Render(header) + count

		case rowStagedFile:
			file := m.staged[row.fileIndex]
			line = m.renderFileLine(file)

		case rowChangesFile:
			file := m.unstaged[row.fileIndex]
			line = m.renderFileLine(file)
		}

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

// renderFileLine renders a single file entry
func (m Model) renderFileLine(file diff.FileEntry) string {
	// Get status style and indicator
	statusStyle := m.statusStyles[file.Status]
	if statusStyle.Value() == "" {
		statusStyle = m.normalStyle
	}

	// Status indicator
	var indicator string
	if file.Staged {
		indicator = file.IndexStatus.String()
	} else {
		indicator = file.WorkStatus.String()
	}

	// File icon
	icon := getFileIcon(file.Path)

	// Just show filename, not full path
	filename := filepath.Base(file.Path)

	// Truncate if needed
	maxLen := m.width - 10 // Account for indent, icon, status
	if len(filename) > maxLen && maxLen > 3 {
		filename = filename[:maxLen-3] + "..."
	}

	// Build line: "    icon filename status"
	return fmt.Sprintf("    %s %s %s",
		m.iconStyle.Render(icon),
		filename,
		statusStyle.Render(indicator),
	)
}
