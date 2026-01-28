package app

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Danny-Dasilva/gdiff/internal/git"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/internal/ui/commit"
	"github.com/Danny-Dasilva/gdiff/internal/ui/commitinput"
	"github.com/Danny-Dasilva/gdiff/internal/ui/diffview"
	"github.com/Danny-Dasilva/gdiff/internal/ui/filetree"
	"github.com/Danny-Dasilva/gdiff/internal/ui/statusbar"
	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

// Model is the root application model
type Model struct {
	// Components
	commitInput commitinput.Model
	fileTree    filetree.Model
	diffView    diffview.Model
	statusBar   statusbar.Model
	commitModal commit.Model

	// State
	files          []diff.FileEntry
	currentFile    string
	focused        types.Pane
	keyMap         types.KeyMap
	showStaged     bool                       // Toggle between staged/unstaged view
	diffCache      map[string][]diff.FileDiff // Cache for diffs
	cancelDiffLoad context.CancelFunc         // Cancel pending diff load

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
		commitInput: commitinput.New(),
		fileTree:    filetree.New(keyMap),
		diffView:    diffview.New(keyMap),
		statusBar:   statusbar.New(keyMap),
		commitModal: commit.New(keyMap),
		focused:     types.PaneFileTree,
		keyMap:      keyMap,
		diffCache:   make(map[string][]diff.FileDiff),
		borderStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		titleStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.statusBar.StartSpinner("Loading status..."),
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

func (m *Model) loadDiff(path string, staged bool) tea.Cmd {
	// Cancel any pending diff load
	if m.cancelDiffLoad != nil {
		m.cancelDiffLoad()
	}

	// Check cache first
	cacheKey := path
	if staged {
		cacheKey = "staged:" + path
	}
	if cached, ok := m.diffCache[cacheKey]; ok {
		m.cancelDiffLoad = nil // No async operation pending
		return func() tea.Msg {
			return types.DiffLoadedMsg{Path: path, Diffs: cached, Err: nil}
		}
	}

	// Create cancellable context for this load
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelDiffLoad = cancel

	return func() tea.Msg {
		diffs, err := git.GetFileDiff(ctx, path, staged)
		// If context was cancelled, return nil to indicate stale result
		if ctx.Err() != nil {
			return nil
		}
		return types.DiffLoadedMsg{Path: path, Diffs: diffs, Err: err}
	}
}

// Large diff threshold (lines)
const largeDiffThreshold = 5000

func (m Model) checkLargeDiff(diffs []diff.FileDiff) bool {
	totalLines := 0
	for _, fd := range diffs {
		for _, hunk := range fd.Hunks {
			totalLines += len(hunk.Lines)
		}
	}
	return totalLines > largeDiffThreshold
}

func (m Model) stageFile(path string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := git.StageFile(ctx, path)
		return types.StageCompleteMsg{Path: path, Err: err}
	}
}

func (m Model) unstageFile(path string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := git.UnstageFile(ctx, path)
		return types.UnstageCompleteMsg{Path: path, Err: err}
	}
}

func (m Model) stageCharacters(path string, hunk diff.Hunk, lineIndex, charStart, charEnd int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := git.StageCharacters(ctx, path, hunk, lineIndex, charStart, charEnd)
		return types.StageCompleteMsg{Path: path, Err: err}
	}
}

func (m Model) doCommit(message string, amend bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		args := []string{"commit", "-m", message}
		if amend {
			args = append(args, "--amend")
		}
		out, err := git.RunGitCommand(ctx, args...)
		if err != nil {
			return types.CommitCompleteMsg{Err: err}
		}
		// Extract commit hash from output (simplified)
		return types.CommitCompleteMsg{Hash: out}
	}
}

func (m Model) doPush(force bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		args := []string{"push"}
		if force {
			args = append(args, "--force-with-lease")
		}
		_, err := git.RunGitCommand(ctx, args...)
		return types.PushCompleteMsg{Err: err}
	}
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle commit modal first if visible
	if m.commitModal.Visible() {
		var cmd tea.Cmd
		m.commitModal, cmd = m.commitModal.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check for commit modal messages
		switch msg := msg.(type) {
		case commit.ConfirmMsg:
			cmds = append(cmds, m.statusBar.StartSpinner("Committing..."))
			cmds = append(cmds, m.doCommit(msg.Message, msg.Amend))
		case commit.CancelMsg:
			m.statusBar.SetMessage("Commit cancelled")
		}

		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		m.commitModal.SetSize(msg.Width, msg.Height)

	case spinner.TickMsg:
		// Forward spinner ticks to statusbar
		cmd := m.statusBar.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

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

		case key.Matches(msg, m.keyMap.Commit):
			m.commitModal.Show(false)
			return m, nil

		case key.Matches(msg, m.keyMap.CommitAmend):
			m.commitModal.Show(true)
			return m, nil

		case key.Matches(msg, m.keyMap.Push):
			return m, tea.Batch(
				m.statusBar.StartSpinner("Pushing..."),
				m.doPush(false),
			)

		case key.Matches(msg, m.keyMap.ForcePush):
			return m, tea.Batch(
				m.statusBar.StartSpinner("Force pushing..."),
				m.doPush(true),
			)

		case key.Matches(msg, m.keyMap.StageItem):
			// Stage selection - could be file, lines, or characters
			if m.focused == types.PaneDiffView && m.diffView.IsInCharMode() {
				// Stage selected characters
				if info := m.diffView.GetCharStagingInfo(); info != nil {
					return m, m.stageCharacters(m.currentFile, info.Hunk, info.HunkLineIndex, info.CharStart, info.CharEnd)
				}
			}

		case key.Matches(msg, m.keyMap.StageFile):
			// Stage selected file
			if m.focused == types.PaneFileTree {
				if f := m.fileTree.SelectedFile(); f != nil {
					return m, m.stageFile(f.Path)
				}
			}

		case key.Matches(msg, m.keyMap.UnstageFile):
			// Unstage selected file
			if m.focused == types.PaneFileTree {
				if f := m.fileTree.SelectedFile(); f != nil {
					return m, m.unstageFile(f.Path)
				}
			}

		case key.Matches(msg, m.keyMap.ToggleStagedView):
			// Toggle between staged and unstaged view
			m.showStaged = !m.showStaged
			if m.showStaged {
				m.statusBar.SetMessage("Showing staged changes")
				m.statusBar.SetMode("STAGED")
			} else {
				m.statusBar.SetMessage("Showing unstaged changes")
				m.statusBar.SetMode("NORMAL")
			}
			// Reload current file diff with new view
			if m.currentFile != "" {
				return m, m.loadDiff(m.currentFile, m.showStaged)
			}
			return m, nil

		// Focus commit input with 'i' (insert mode)
		case msg.String() == "i":
			m.focused = types.PaneCommitInput
			m.updateLayout()
			return m, m.commitInput.Focus()
		}

		// Delegate to focused component
		switch m.focused {
		case types.PaneCommitInput:
			// Handle commit input
			switch msg.String() {
			case "enter":
				// Commit with the message
				if msg := m.commitInput.Value(); msg != "" {
					cmds = append(cmds, m.statusBar.StartSpinner("Committing..."))
					cmds = append(cmds, m.doCommit(msg, false))
					m.commitInput.Reset()
					m.focused = types.PaneFileTree
					m.updateLayout()
				}
			case "esc":
				// Exit commit input
				m.focused = types.PaneFileTree
				m.updateLayout()
			default:
				var cmd tea.Cmd
				m.commitInput, cmd = m.commitInput.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}

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
		m.statusBar.StopSpinner()
		if msg.Err != nil {
			m.statusBar.SetMessage("Error: " + msg.Err.Error())
		} else {
			m.files = msg.Files
			m.fileTree.SetFiles(msg.Files)
			m.updateCounts()

			// Load first file's diff
			if len(msg.Files) > 0 {
				f := msg.Files[0]
				cmds = append(cmds, m.statusBar.StartSpinner("Loading diff..."))
				cmds = append(cmds, m.loadDiff(f.Path, f.Staged))
			}
		}

	case types.DiffLoadedMsg:
		m.statusBar.StopSpinner()
		if msg.Err != nil {
			m.statusBar.SetMessage("Error loading diff: " + msg.Err.Error())
		} else {
			m.currentFile = msg.Path
			// Cache the diff
			cacheKey := msg.Path
			if m.showStaged {
				cacheKey = "staged:" + msg.Path
			}
			m.diffCache[cacheKey] = msg.Diffs

			// Warn about large diffs
			if m.checkLargeDiff(msg.Diffs) {
				m.statusBar.SetMessage("Warning: Large diff - character highlighting disabled")
			}
			m.diffView.SetDiff(msg.Path, msg.Diffs)
		}

	case types.FileSelectedMsg:
		cmds = append(cmds, m.statusBar.StartSpinner("Loading diff..."))
		cmds = append(cmds, m.loadDiff(msg.Path, msg.Staged))

	case types.StageCompleteMsg:
		if msg.Err != nil {
			m.statusBar.SetMessage("Stage error: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage("Staged: " + msg.Path)
			// Invalidate cache for this file
			delete(m.diffCache, msg.Path)
			delete(m.diffCache, "staged:"+msg.Path)
			cmds = append(cmds, m.loadStatus())
		}

	case types.UnstageCompleteMsg:
		if msg.Err != nil {
			m.statusBar.SetMessage("Unstage error: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage("Unstaged: " + msg.Path)
			// Invalidate cache for this file
			delete(m.diffCache, msg.Path)
			delete(m.diffCache, "staged:"+msg.Path)
			cmds = append(cmds, m.loadStatus())
		}

	case types.CommitCompleteMsg:
		m.statusBar.StopSpinner()
		if msg.Err != nil {
			m.statusBar.SetMessage("Commit error: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage("Committed successfully")
			// Clear entire cache after commit
			m.diffCache = make(map[string][]diff.FileDiff)
			cmds = append(cmds, m.loadStatus())
		}

	case types.PushCompleteMsg:
		m.statusBar.StopSpinner()
		if msg.Err != nil {
			m.statusBar.SetMessage("Push error: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage("Pushed successfully")
		}

	case branchLoadedMsg:
		m.statusBar.SetBranch(msg.branch)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) switchFocus() {
	// Cycle: FileTree -> DiffView -> FileTree (skip commit input in Tab cycle)
	switch m.focused {
	case types.PaneCommitInput:
		m.focused = types.PaneFileTree
	case types.PaneFileTree:
		m.focused = types.PaneDiffView
	case types.PaneDiffView:
		m.focused = types.PaneFileTree
	}
	m.updateLayout()
}

func (m *Model) updateLayout() {
	// Reserve space for status bar and commit input
	statusHeight := m.statusBar.HelpHeight()
	commitInputHeight := m.commitInput.Height()
	contentHeight := m.height - statusHeight - commitInputHeight - 2 // borders

	// Split width: 30% file tree, 70% diff view
	fileTreeWidth := m.width * 30 / 100
	fileTreeWidth = max(fileTreeWidth, 20)
	diffViewWidth := m.width - fileTreeWidth - 3 // border + separator

	m.commitInput.SetWidth(fileTreeWidth - 2)
	m.fileTree.SetSize(fileTreeWidth-2, contentHeight-2)
	m.diffView.SetSize(diffViewWidth-2, contentHeight+commitInputHeight)
	m.statusBar.SetWidth(m.width)

	// Update focus states
	m.commitInput.Blur()
	m.fileTree.SetFocused(false)
	m.diffView.SetFocused(false)

	switch m.focused {
	case types.PaneCommitInput:
		m.commitInput.Focus()
	case types.PaneFileTree:
		m.fileTree.SetFocused(true)
	case types.PaneDiffView:
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

	// Overlay commit modal if visible (render on top)
	if m.commitModal.Visible() {
		return m.commitModal.View()
	}

	// Catppuccin Mocha colors
	base := lipgloss.Color("#1e1e2e")
	surface := lipgloss.Color("#313244")
	text := lipgloss.Color("#cdd6f4")
	subtext := lipgloss.Color("#a6adc8")
	blue := lipgloss.Color("#89b4fa")
	mauve := lipgloss.Color("#cba6f7")

	// Outer frame dimensions (reserve space for frame border)
	frameWidth := m.width - 2
	frameHeight := m.height - 2

	// Status bar height
	statusHeight := m.statusBar.HelpHeight()

	// Inner content dimensions (inside frame, minus title bar and status bar)
	titleBarHeight := 1
	innerHeight := frameHeight - titleBarHeight - statusHeight - 1
	innerWidth := frameWidth - 2

	// Panel widths
	fileTreeWidth := innerWidth * 30 / 100
	fileTreeWidth = max(fileTreeWidth, 20)
	diffViewWidth := innerWidth - fileTreeWidth - 1

	// Commit input section
	commitInputHeight := m.commitInput.Height()
	commitInputView := m.commitInput.View()

	// File tree section (below commit input)
	fileTreeHeight := innerHeight - commitInputHeight - 3
	fileTreeBorderColor := surface
	if m.focused == types.PaneFileTree {
		fileTreeBorderColor = blue
	}
	fileTreeBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(fileTreeBorderColor)

	fileTreePane := fileTreeBorder.
		Width(fileTreeWidth - 2).
		Height(fileTreeHeight).
		Render(m.fileTree.View())

	// Combine commit input and file tree vertically
	leftPane := lipgloss.JoinVertical(lipgloss.Left, commitInputView, fileTreePane)

	// Build diff view pane (full height)
	diffViewBorderColor := surface
	if m.focused == types.PaneDiffView {
		diffViewBorderColor = blue
	}
	diffViewBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(diffViewBorderColor)

	diffTitle := " Diff "
	if m.currentFile != "" {
		diffTitle = " " + m.currentFile + " "
	}
	diffTitleStyled := lipgloss.NewStyle().
		Bold(true).
		Foreground(text).
		Render(diffTitle)

	diffViewPane := diffViewBorder.
		Width(diffViewWidth - 2).
		Height(innerHeight).
		Render(diffTitleStyled + "\n" + m.diffView.View())

	// Combine panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, diffViewPane)

	// Title bar
	titleText := " gdiff "
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(base).
		Background(mauve).
		Padding(0, 1)

	subtitleText := " Git Diff TUI "
	subtitleStyle := lipgloss.NewStyle().
		Foreground(subtext).
		Background(surface).
		Padding(0, 1)

	titleBar := titleStyle.Render(titleText) + subtitleStyle.Render(subtitleText)
	titleBarPadding := frameWidth - lipgloss.Width(titleBar) - 2
	if titleBarPadding > 0 {
		titleBarBg := lipgloss.NewStyle().Background(surface)
		titleBar = titleBar + titleBarBg.Render(strings.Repeat(" ", titleBarPadding))
	}

	// Status bar (set width to fit inside frame)
	m.statusBar.SetWidth(frameWidth - 2)
	statusBar := m.statusBar.View()

	// Combine title bar, content, and status bar
	innerContent := lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		content,
		statusBar,
	)

	// Outer frame with rounded border
	outerFrame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(surface).
		Width(frameWidth).
		Height(frameHeight)

	return outerFrame.Render(innerContent)
}
