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
	statusStyles  map[diff.FileStatus]lipgloss.Style

	// File type icon styles (color-coded)
	iconStyles map[string]lipgloss.Style
}

// New creates a new file tree model
func New(keyMap types.KeyMap) Model {
	return Model{
		keyMap:        keyMap,
		normalStyle:   lipgloss.NewStyle(),
		selectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("238")),
		focusedStyle:  lipgloss.NewStyle().Background(lipgloss.Color("62")),
		headerStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")),
		countStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true),
		statusStyles: map[diff.FileStatus]lipgloss.Style{
			diff.StatusModified:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true), // Orange
			diff.StatusAdded:     lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true),  // Green
			diff.StatusDeleted:   lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Bold(true), // Red
			diff.StatusRenamed:   lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true), // Purple
			diff.StatusUntracked: lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true),  // Green (new file)
			diff.StatusUnmerged:  lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true), // Orange
		},
		// Color-coded file type icons
		iconStyles: map[string]lipgloss.Style{
			"go":         lipgloss.NewStyle().Foreground(lipgloss.Color("81")),  // Cyan/Go blue
			"js":         lipgloss.NewStyle().Foreground(lipgloss.Color("226")), // Yellow
			"ts":         lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
			"py":         lipgloss.NewStyle().Foreground(lipgloss.Color("220")), // Yellow/Gold
			"rust":       lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
			"ruby":       lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
			"java":       lipgloss.NewStyle().Foreground(lipgloss.Color("166")), // Orange/Red
			"c":          lipgloss.NewStyle().Foreground(lipgloss.Color("75")),  // Blue
			"cpp":        lipgloss.NewStyle().Foreground(lipgloss.Color("204")), // Pink
			"json":       lipgloss.NewStyle().Foreground(lipgloss.Color("185")), // Yellow
			"yaml":       lipgloss.NewStyle().Foreground(lipgloss.Color("167")), // Red/Orange
			"toml":       lipgloss.NewStyle().Foreground(lipgloss.Color("209")), // Orange
			"md":         lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
			"html":       lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
			"css":        lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
			"shell":      lipgloss.NewStyle().Foreground(lipgloss.Color("113")), // Green
			"sql":        lipgloss.NewStyle().Foreground(lipgloss.Color("75")),  // Blue
			"docker":     lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
			"git":        lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
			"config":     lipgloss.NewStyle().Foreground(lipgloss.Color("243")), // Gray
			"lock":       lipgloss.NewStyle().Foreground(lipgloss.Color("243")), // Gray
			"test":       lipgloss.NewStyle().Foreground(lipgloss.Color("113")), // Green
			"proto":      lipgloss.NewStyle().Foreground(lipgloss.Color("75")),  // Blue
			"graphql":    lipgloss.NewStyle().Foreground(lipgloss.Color("200")), // Magenta
			"vue":        lipgloss.NewStyle().Foreground(lipgloss.Color("113")), // Green
			"svelte":     lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
			"kotlin":     lipgloss.NewStyle().Foreground(lipgloss.Color("99")),  // Purple
			"swift":      lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
			"php":        lipgloss.NewStyle().Foreground(lipgloss.Color("99")),  // Purple
			"elixir":     lipgloss.NewStyle().Foreground(lipgloss.Color("99")),  // Purple
			"erlang":     lipgloss.NewStyle().Foreground(lipgloss.Color("161")), // Magenta
			"haskell":    lipgloss.NewStyle().Foreground(lipgloss.Color("99")),  // Purple
			"lua":        lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
			"vim":        lipgloss.NewStyle().Foreground(lipgloss.Color("113")), // Green
			"default":    lipgloss.NewStyle().Foreground(lipgloss.Color("246")), // Default gray
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

// fileIconInfo contains icon and style type for a file
type fileIconInfo struct {
	icon      string
	styleType string
}

// getFileIcon returns an icon and style type based on file extension/name
// Uses simple Unicode characters that work in any terminal
func getFileIcon(path string) fileIconInfo {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	// Check for special filenames first
	switch base {
	case "dockerfile", "containerfile":
		return fileIconInfo{"◈", "docker"}
	case "makefile", "gnumakefile":
		return fileIconInfo{"◇", "config"}
	case ".gitignore", ".gitattributes", ".gitmodules":
		return fileIconInfo{"●", "git"}
	case ".env", ".env.local", ".env.example":
		return fileIconInfo{"◇", "config"}
	case "package.json":
		return fileIconInfo{"◆", "json"}
	case "tsconfig.json", "jsconfig.json":
		return fileIconInfo{"◆", "ts"}
	case "cargo.toml", "cargo.lock":
		return fileIconInfo{"◆", "rust"}
	case "go.mod", "go.sum":
		return fileIconInfo{"◆", "go"}
	case "requirements.txt", "pyproject.toml", "setup.py":
		return fileIconInfo{"◆", "py"}
	case "gemfile", "gemfile.lock":
		return fileIconInfo{"◆", "ruby"}
	case "yarn.lock", "package-lock.json", "pnpm-lock.yaml":
		return fileIconInfo{"○", "lock"}
	case "license", "license.md", "license.txt":
		return fileIconInfo{"◇", "config"}
	case "readme.md", "readme.txt", "readme":
		return fileIconInfo{"◆", "md"}
	}

	// Check by extension
	switch ext {
	// Go
	case ".go":
		return fileIconInfo{"◆", "go"}
	// JavaScript/TypeScript
	case ".js", ".mjs", ".cjs":
		return fileIconInfo{"◆", "js"}
	case ".jsx":
		return fileIconInfo{"◆", "js"}
	case ".ts", ".mts", ".cts":
		return fileIconInfo{"◆", "ts"}
	case ".tsx":
		return fileIconInfo{"◆", "ts"}
	// Python
	case ".py", ".pyw", ".pyi":
		return fileIconInfo{"◆", "py"}
	case ".ipynb":
		return fileIconInfo{"◆", "py"}
	// Rust
	case ".rs":
		return fileIconInfo{"◆", "rust"}
	// Ruby
	case ".rb", ".erb":
		return fileIconInfo{"◆", "ruby"}
	case ".rake":
		return fileIconInfo{"◆", "ruby"}
	// Java/JVM
	case ".java":
		return fileIconInfo{"◆", "java"}
	case ".kt", ".kts":
		return fileIconInfo{"◆", "kotlin"}
	case ".scala":
		return fileIconInfo{"◆", "java"}
	case ".groovy":
		return fileIconInfo{"◆", "java"}
	// C/C++
	case ".c", ".h":
		return fileIconInfo{"◆", "c"}
	case ".cpp", ".hpp", ".cc", ".cxx", ".hxx":
		return fileIconInfo{"◆", "cpp"}
	// C#/F#
	case ".cs":
		return fileIconInfo{"◆", "cpp"}
	case ".fs", ".fsx":
		return fileIconInfo{"◆", "cpp"}
	// Swift/Objective-C
	case ".swift":
		return fileIconInfo{"◆", "swift"}
	case ".m", ".mm":
		return fileIconInfo{"◆", "c"}
	// PHP
	case ".php":
		return fileIconInfo{"◆", "php"}
	// Data formats
	case ".json", ".jsonc":
		return fileIconInfo{"◇", "json"}
	case ".yaml", ".yml":
		return fileIconInfo{"◇", "yaml"}
	case ".toml":
		return fileIconInfo{"◇", "toml"}
	case ".xml", ".plist":
		return fileIconInfo{"◇", "html"}
	case ".csv":
		return fileIconInfo{"◇", "config"}
	// Markup/Docs
	case ".md", ".markdown", ".mdx":
		return fileIconInfo{"◆", "md"}
	case ".rst":
		return fileIconInfo{"◆", "md"}
	case ".txt":
		return fileIconInfo{"○", "default"}
	case ".pdf":
		return fileIconInfo{"○", "default"}
	// Web
	case ".html", ".htm":
		return fileIconInfo{"◆", "html"}
	case ".css":
		return fileIconInfo{"◆", "css"}
	case ".scss", ".sass":
		return fileIconInfo{"◆", "css"}
	case ".less":
		return fileIconInfo{"◆", "css"}
	case ".vue":
		return fileIconInfo{"◆", "vue"}
	case ".svelte":
		return fileIconInfo{"◆", "svelte"}
	// Shell
	case ".sh", ".bash", ".zsh", ".fish":
		return fileIconInfo{"◆", "shell"}
	case ".ps1", ".psm1":
		return fileIconInfo{"◆", "shell"}
	// Database
	case ".sql":
		return fileIconInfo{"◆", "sql"}
	case ".prisma":
		return fileIconInfo{"◆", "sql"}
	// GraphQL/Proto
	case ".graphql", ".gql":
		return fileIconInfo{"◆", "graphql"}
	case ".proto":
		return fileIconInfo{"◆", "proto"}
	// Functional languages
	case ".ex", ".exs":
		return fileIconInfo{"◆", "elixir"}
	case ".erl", ".hrl":
		return fileIconInfo{"◆", "erlang"}
	case ".hs", ".lhs":
		return fileIconInfo{"◆", "haskell"}
	case ".clj", ".cljs", ".cljc":
		return fileIconInfo{"◆", "haskell"}
	case ".ml", ".mli":
		return fileIconInfo{"◆", "haskell"}
	// Lua
	case ".lua":
		return fileIconInfo{"◆", "lua"}
	// Vim
	case ".vim":
		return fileIconInfo{"◆", "vim"}
	// Config files
	case ".conf", ".cfg", ".ini":
		return fileIconInfo{"◇", "config"}
	case ".env":
		return fileIconInfo{"◇", "config"}
	// Docker
	case ".dockerfile":
		return fileIconInfo{"◈", "docker"}
	// Git
	case ".gitignore", ".gitattributes":
		return fileIconInfo{"●", "git"}
	// Images
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".svg":
		return fileIconInfo{"○", "default"}
	// Archives
	case ".zip", ".tar", ".gz", ".rar", ".7z":
		return fileIconInfo{"○", "default"}
	// Binaries
	case ".exe", ".dll", ".so", ".dylib":
		return fileIconInfo{"○", "default"}
	// Tests (by naming convention)
	default:
		// Check for test files
		if strings.Contains(base, "_test.") || strings.Contains(base, ".test.") ||
			strings.Contains(base, ".spec.") || strings.HasPrefix(base, "test_") {
			return fileIconInfo{"●", "test"}
		}
		return fileIconInfo{"○", "default"}
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
			// Staged header with checkmark icon and green accent
			stagedIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true).Render("")
			header := fmt.Sprintf(" %s %s STAGED", chevron, stagedIcon)
			count := m.countStyle.Render(fmt.Sprintf(" (%d)", len(m.staged)))
			line = m.headerStyle.Render(header) + count

		case rowChangesHeader:
			chevron := "▼"
			if m.unstagedCollapsed {
				chevron = "▶"
			}
			// Changes header with pencil icon and orange accent
			changesIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("")
			header := fmt.Sprintf(" %s %s CHANGES", chevron, changesIcon)
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

	// File icon with color
	iconInfo := getFileIcon(file.Path)
	iconStyle := m.iconStyles[iconInfo.styleType]
	if iconStyle.Value() == "" {
		iconStyle = m.iconStyles["default"]
	}

	// Just show filename, not full path
	filename := filepath.Base(file.Path)

	// Truncate if needed
	maxLen := m.width - 10 // Account for indent, icon, status
	if len(filename) > maxLen && maxLen > 3 {
		filename = filename[:maxLen-3] + "..."
	}

	// Build line: "    icon filename status"
	return fmt.Sprintf("    %s %s %s",
		iconStyle.Render(iconInfo.icon),
		filename,
		statusStyle.Render(indicator),
	)
}
