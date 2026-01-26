package app

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/danny/gdiff/internal/git"
	"github.com/danny/gdiff/internal/types"
	"github.com/danny/gdiff/internal/ui/diffview"
	"github.com/danny/gdiff/internal/ui/filetree"
	"github.com/danny/gdiff/internal/ui/statusbar"
	"github.com/danny/gdiff/pkg/diff"
)

// Model is the root application model
type Model struct {
	// Components
	fileTree  filetree.Model
	diffView  diffview.Model
	statusBar statusbar.Model

	// State
	files       []diff.FileEntry
	currentFile string
	focused     types.Pane
	keyMap      types.KeyMap

	// Layout
	width  int
	height int

	// Styles
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
}

// New creates a new application model
func New() Model {
	keyMap := types.DefaultKeyMap()

	return Model{
		fileTree:    filetree.New(keyMap),
		diffView:    diffview.New(keyMap),
		statusBar:   statusbar.New(keyMap),
		focused:     types.PaneFileTree,
		keyMap:      keyMap,
		borderStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		titleStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadStatus(),
		m.loadBranch(),
	)
}

func (m Model) loadStatus() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		files, err := git.GetStatus(ctx)
		return types.StatusLoadedMsg{Files: files, Err: err}
	}
}

func (m Model) loadBranch() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		branch, _ := git.GetCurrentBranch(ctx)
		return branchLoadedMsg{branch: branch}
	}
}

type branchLoadedMsg struct {
	branch string
}

func (m Model) loadDiff(path string, staged bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		diffs, err := git.GetFileDiff(ctx, path, staged)
		return types.DiffLoadedMsg{Path: path, Diffs: diffs, Err: err}
	}
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.KeyMsg:
		// Global key handling
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.statusBar.ToggleHelp()
			m.updateLayout()
			return m, nil

		case key.Matches(msg, m.keyMap.SwitchPane):
			m.switchFocus()
			return m, nil
		}

		// Delegate to focused component
		switch m.focused {
		case types.PaneFileTree:
			var cmd tea.Cmd
			m.fileTree, cmd = m.fileTree.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

		case types.PaneDiffView:
			var cmd tea.Cmd
			m.diffView, cmd = m.diffView.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case types.StatusLoadedMsg:
		if msg.Err != nil {
			m.statusBar.SetMessage("Error: " + msg.Err.Error())
		} else {
			m.files = msg.Files
			m.fileTree.SetFiles(msg.Files)
			m.updateCounts()

			// Load first file's diff
			if len(msg.Files) > 0 {
				f := msg.Files[0]
				cmds = append(cmds, m.loadDiff(f.Path, f.Staged))
			}
		}

	case types.DiffLoadedMsg:
		if msg.Err != nil {
			m.statusBar.SetMessage("Error loading diff: " + msg.Err.Error())
		} else {
			m.currentFile = msg.Path
			m.diffView.SetDiff(msg.Path, msg.Diffs)
		}

	case types.FileSelectedMsg:
		cmds = append(cmds, m.loadDiff(msg.Path, msg.Staged))

	case branchLoadedMsg:
		m.statusBar.SetBranch(msg.branch)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) switchFocus() {
	switch m.focused {
	case types.PaneFileTree:
		m.focused = types.PaneDiffView
		m.fileTree.SetFocused(false)
		m.diffView.SetFocused(true)
	case types.PaneDiffView:
		m.focused = types.PaneFileTree
		m.fileTree.SetFocused(true)
		m.diffView.SetFocused(false)
	}
}

func (m *Model) updateLayout() {
	// Reserve space for status bar
	statusHeight := m.statusBar.HelpHeight()
	contentHeight := m.height - statusHeight - 2 // borders

	// Split width: 30% file tree, 70% diff view
	fileTreeWidth := m.width * 30 / 100
	if fileTreeWidth < 20 {
		fileTreeWidth = 20
	}
	diffViewWidth := m.width - fileTreeWidth - 3 // border + separator

	m.fileTree.SetSize(fileTreeWidth-2, contentHeight)
	m.diffView.SetSize(diffViewWidth-2, contentHeight)
	m.statusBar.SetWidth(m.width)

	// Set initial focus
	if m.focused == types.PaneFileTree {
		m.fileTree.SetFocused(true)
		m.diffView.SetFocused(false)
	} else {
		m.fileTree.SetFocused(false)
		m.diffView.SetFocused(true)
	}
}

func (m *Model) updateCounts() {
	total := len(m.files)
	staged := 0
	for _, f := range m.files {
		if f.Staged {
			staged++
		}
	}
	m.statusBar.SetCounts(total, staged)
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate dimensions
	statusHeight := m.statusBar.HelpHeight()
	contentHeight := m.height - statusHeight

	fileTreeWidth := m.width * 30 / 100
	if fileTreeWidth < 20 {
		fileTreeWidth = 20
	}
	diffViewWidth := m.width - fileTreeWidth - 1

	// Build file tree pane
	fileTreeBorder := m.borderStyle
	if m.focused == types.PaneFileTree {
		fileTreeBorder = fileTreeBorder.BorderForeground(lipgloss.Color("62"))
	}
	fileTreePane := fileTreeBorder.
		Width(fileTreeWidth - 2).
		Height(contentHeight - 2).
		Render(m.titleStyle.Render(" Files ") + "\n" + m.fileTree.View())

	// Build diff view pane
	diffViewBorder := m.borderStyle
	if m.focused == types.PaneDiffView {
		diffViewBorder = diffViewBorder.BorderForeground(lipgloss.Color("62"))
	}
	diffTitle := " Diff "
	if m.currentFile != "" {
		diffTitle = " " + m.currentFile + " "
	}
	diffViewPane := diffViewBorder.
		Width(diffViewWidth - 2).
		Height(contentHeight - 2).
		Render(m.titleStyle.Render(diffTitle) + "\n" + m.diffView.View())

	// Combine panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, fileTreePane, diffViewPane)

	// Add status bar
	return lipgloss.JoinVertical(lipgloss.Left, content, m.statusBar.View())
}
