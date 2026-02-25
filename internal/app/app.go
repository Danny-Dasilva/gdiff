package app

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Danny-Dasilva/gdiff/internal/git"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/internal/ui/commit"
	"github.com/Danny-Dasilva/gdiff/internal/ui/commitinput"
	"github.com/Danny-Dasilva/gdiff/internal/ui/diffview"
	"github.com/Danny-Dasilva/gdiff/internal/ui/filetree"
	"github.com/Danny-Dasilva/gdiff/internal/ui/helpoverlay"
	"github.com/Danny-Dasilva/gdiff/internal/ui/statusbar"
	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

type Model struct {
	commitInput commitinput.Model
	fileTree    filetree.Model
	diffView    diffview.Model
	statusBar   statusbar.Model
	commitModal commit.Model
	helpOverlay helpoverlay.Model

	files          []diff.FileEntry
	currentFile    string
	focused        types.Pane
	keyMap         types.KeyMap
	showStaged     bool
	diffCache      map[string][]diff.FileDiff
	cancelDiffLoad context.CancelFunc

	width            int
	height           int
	sidebarCollapsed bool

	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
}

func New(colorblind bool) Model {
	keyMap := types.DefaultKeyMap()

	return Model{
		commitInput: commitinput.New(),
		fileTree:    filetree.New(keyMap),
		diffView:    diffview.New(keyMap, colorblind),
		statusBar:   statusbar.New(keyMap),
		commitModal: commit.New(keyMap),
		helpOverlay: helpoverlay.New(),
		focused:     types.PaneFileTree,
		keyMap:      keyMap,
		diffCache:   make(map[string][]diff.FileDiff),
		borderStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		titleStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
	}
}

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

func diffCacheKey(path string, staged bool) string {
	if staged {
		return "staged:" + path
	}
	return path
}

func (m *Model) invalidateFileCache(path string) {
	delete(m.diffCache, path)
	delete(m.diffCache, "staged:"+path)
}

func (m *Model) loadDiff(path string, staged bool) tea.Cmd {
	if m.cancelDiffLoad != nil {
		m.cancelDiffLoad()
	}

	cacheKey := diffCacheKey(path, staged)
	if cached, ok := m.diffCache[cacheKey]; ok {
		m.cancelDiffLoad = nil
		return func() tea.Msg {
			return types.DiffLoadedMsg{Path: path, Diffs: cached, Err: nil}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelDiffLoad = cancel

	return func() tea.Msg {
		diffs, err := git.GetFileDiff(ctx, path, staged)
		if ctx.Err() != nil {
			return nil
		}
		return types.DiffLoadedMsg{Path: path, Diffs: diffs, Err: err}
	}
}

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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.helpOverlay.Visible() {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.helpOverlay.SetSize(msg.Width, msg.Height)
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, m.keyMap.Help), key.Matches(msg, m.keyMap.Escape):
				m.helpOverlay.Hide()
			case key.Matches(msg, m.keyMap.Quit):
				return m, tea.Quit
			}
		}
		return m, nil
	}

	if m.commitModal.Visible() {
		var cmd tea.Cmd
		m.commitModal, cmd = m.commitModal.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

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
		m.helpOverlay.SetSize(msg.Width, msg.Height)

	case spinner.TickMsg:
		cmd := m.statusBar.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			frameX := 1
			frameY := 1
			titleBarHeight := 1
			innerWidth := m.width - 2
			fileTreeWidth := innerWidth * 30 / 100
			if fileTreeWidth < 20 {
				fileTreeWidth = 20
			}
			if m.sidebarCollapsed {
				fileTreeWidth = 0
			}
			commitInputHeight := m.commitInput.Height()

			mouse := msg.Mouse()
			clickX := mouse.X - frameX
			clickY := mouse.Y - frameY - titleBarHeight

			if clickY >= 0 {
				if !m.sidebarCollapsed && clickX >= 0 && clickX < fileTreeWidth {
					if clickY < commitInputHeight {
						m.focused = types.PaneCommitInput
					} else {
						m.focused = types.PaneFileTree
						adjustedY := clickY - commitInputHeight - 1
						clickMsg := tea.MouseClickMsg{X: mouse.X, Y: adjustedY, Button: msg.Button}
						m.fileTree.SetFocused(true)
						var cmd tea.Cmd
						m.fileTree, cmd = m.fileTree.Update(clickMsg)
						if cmd != nil {
							cmds = append(cmds, cmd)
						}
					}
				} else if clickX >= fileTreeWidth {
					m.focused = types.PaneDiffView
				}
				m.updateLayout()
			}
		}
		return m, tea.Batch(cmds...)

	case tea.MouseWheelMsg:
		innerWidth := m.width - 2
		fileTreeWidth := innerWidth * 30 / 100
		if fileTreeWidth < 20 {
			fileTreeWidth = 20
		}
		mouse := msg.Mouse()
		clickX := mouse.X - 1

		if clickX >= 0 && clickX < fileTreeWidth {
			fileTreeFocused := m.focused == types.PaneFileTree
			if !fileTreeFocused {
				m.fileTree.SetFocused(true)
			}
			var cmd tea.Cmd
			m.fileTree, cmd = m.fileTree.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			if !fileTreeFocused {
				m.fileTree.SetFocused(false)
			}
		} else if clickX >= fileTreeWidth {
			diffFocused := m.focused == types.PaneDiffView
			if !diffFocused {
				m.diffView.SetFocused(true)
			}
			var cmd tea.Cmd
			m.diffView, cmd = m.diffView.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			if !diffFocused {
				m.diffView.SetFocused(false)
			}
		}
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.helpOverlay.SetSize(m.width, m.height)
			m.helpOverlay.Toggle()
			return m, nil

		case key.Matches(msg, m.keyMap.SwitchPane):
			m.switchFocus()
			return m, nil

		case key.Matches(msg, m.keyMap.ToggleSidebar):
			m.sidebarCollapsed = !m.sidebarCollapsed
			if m.sidebarCollapsed && m.focused != types.PaneDiffView {
				m.focused = types.PaneDiffView
			}
			m.updateLayout()
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
			if m.focused == types.PaneDiffView && m.diffView.IsInCharMode() {
				if info := m.diffView.GetCharStagingInfo(); info != nil {
					return m, m.stageCharacters(m.currentFile, info.Hunk, info.HunkLineIndex, info.CharStart, info.CharEnd)
				}
			}

		case key.Matches(msg, m.keyMap.StageFile):
			if m.focused == types.PaneFileTree {
				if f := m.fileTree.SelectedFile(); f != nil {
					return m, m.stageFile(f.Path)
				}
			}

		case key.Matches(msg, m.keyMap.UnstageFile):
			if m.focused == types.PaneFileTree {
				if f := m.fileTree.SelectedFile(); f != nil {
					return m, m.unstageFile(f.Path)
				}
			}

		case key.Matches(msg, m.keyMap.ToggleStagedView):
			m.showStaged = !m.showStaged
			if m.showStaged {
				m.statusBar.SetMessage("Showing staged changes")
				m.statusBar.SetMode("STAGED")
			} else {
				m.statusBar.SetMessage("Showing unstaged changes")
				m.statusBar.SetMode("NORMAL")
			}
			if m.currentFile != "" {
				return m, m.loadDiff(m.currentFile, m.showStaged)
			}
			return m, nil

		case msg.String() == "i":
			m.focused = types.PaneCommitInput
			m.updateLayout()
			return m, m.commitInput.Focus()
		}

		switch m.focused {
		case types.PaneCommitInput:
			switch msg.String() {
			case "enter":
				if msg := m.commitInput.Value(); msg != "" {
					cmds = append(cmds, m.statusBar.StartSpinner("Committing..."))
					cmds = append(cmds, m.doCommit(msg, false))
					m.commitInput.Reset()
					m.focused = types.PaneFileTree
					m.updateLayout()
				}
			case "esc":
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
			m.diffCache[diffCacheKey(msg.Path, m.showStaged)] = msg.Diffs

			if m.checkLargeDiff(msg.Diffs) {
				m.statusBar.SetMessage("Warning: Large diff - character highlighting disabled")
			}
			m.diffView.SetDiff(msg.Path, msg.Diffs)
		}

	case types.FocusChangedMsg:
		switch msg.Pane {
		case types.PaneDiffView:
			m.focused = types.PaneDiffView
			m.fileTree.SetFocused(false)
			m.diffView.SetFocused(true)
		case types.PaneFileTree:
			m.focused = types.PaneFileTree
			m.fileTree.SetFocused(true)
			m.diffView.SetFocused(false)
		}
		m.statusBar.SetFocusedPane(m.focused)

	case types.SpaceToggleMsg:
		if msg.Staged {
			cmds = append(cmds, m.unstageFile(msg.Path))
		} else {
			cmds = append(cmds, m.stageFile(msg.Path))
		}

	case types.FileSelectedMsg:
		cmds = append(cmds, m.statusBar.StartSpinner("Loading diff..."))
		cmds = append(cmds, m.loadDiff(msg.Path, msg.Staged))

	case types.StageCompleteMsg:
		if msg.Err != nil {
			m.statusBar.SetMessage("Stage error: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage("Staged: " + msg.Path)
			m.invalidateFileCache(msg.Path)
			cmds = append(cmds, m.loadStatus())
		}

	case types.UnstageCompleteMsg:
		if msg.Err != nil {
			m.statusBar.SetMessage("Unstage error: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage("Unstaged: " + msg.Path)
			m.invalidateFileCache(msg.Path)
			cmds = append(cmds, m.loadStatus())
		}

	case types.CommitCompleteMsg:
		m.statusBar.StopSpinner()
		if msg.Err != nil {
			m.statusBar.SetMessage("Commit error: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage("Committed successfully")
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
	frameWidth := m.width - 2
	frameHeight := m.height - 2

	statusHeight := m.statusBar.HelpHeight()
	titleBarHeight := 1
	panelHeight := frameHeight - titleBarHeight - statusHeight

	innerWidth := frameWidth
	var fileTreeWidth, diffViewWidth int
	if m.sidebarCollapsed {
		fileTreeWidth = 0
		diffViewWidth = innerWidth
	} else {
		fileTreeWidth = innerWidth * 30 / 100
		fileTreeWidth = max(fileTreeWidth, 20)
		diffViewWidth = innerWidth - fileTreeWidth - 1
	}

	commitInputHeight := m.commitInput.Height()
	fileTreeContentHeight := panelHeight - commitInputHeight - 2
	if fileTreeContentHeight < 1 {
		fileTreeContentHeight = 1
	}
	diffViewContentHeight := panelHeight - 2
	if diffViewContentHeight < 1 {
		diffViewContentHeight = 1
	}

	if !m.sidebarCollapsed {
		m.commitInput.SetWidth(fileTreeWidth - 2)
		m.fileTree.SetSize(fileTreeWidth-2, fileTreeContentHeight)
	}
	m.diffView.SetSize(diffViewWidth-2, diffViewContentHeight-1)
	m.statusBar.SetWidth(frameWidth - 2)

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
	m.statusBar.SetFocusedPane(m.focused)
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

func (m Model) newView(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m Model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return m.newView("Loading...")
	}

	if m.helpOverlay.Visible() {
		return m.newView(m.helpOverlay.View())
	}
	if m.commitModal.Visible() {
		return m.newView(m.commitModal.View())
	}

	base := lipgloss.Color("#1e1e2e")
	surface := lipgloss.Color("#313244")
	text := lipgloss.Color("#cdd6f4")
	subtext := lipgloss.Color("#a6adc8")
	green := lipgloss.Color("#a6e3a1")
	mauve := lipgloss.Color("#cba6f7")

	frameWidth := m.width - 2
	frameHeight := m.height - 2

	statusHeight := m.statusBar.HelpHeight()
	titleBarHeight := 1
	panelHeight := frameHeight - titleBarHeight - statusHeight
	innerWidth := frameWidth

	var fileTreeWidth, diffViewWidth int
	if m.sidebarCollapsed {
		fileTreeWidth = 0
		diffViewWidth = innerWidth
	} else {
		fileTreeWidth = innerWidth * 30 / 100
		fileTreeWidth = max(fileTreeWidth, 20)
		diffViewWidth = innerWidth - fileTreeWidth - 1
	}

	diffViewContentHeight := panelHeight - 2
	if diffViewContentHeight < 1 {
		diffViewContentHeight = 1
	}
	diffViewBorderColor := surface
	if m.focused == types.PaneDiffView {
		diffViewBorderColor = green
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
		Height(diffViewContentHeight).
		Render(diffTitleStyled + "\n" + m.diffView.View())

	var content string
	if m.sidebarCollapsed {
		content = diffViewPane
	} else {
		commitInputHeight := m.commitInput.Height()
		commitInputView := m.commitInput.View()

		fileTreeContentHeight := panelHeight - commitInputHeight - 2
		if fileTreeContentHeight < 1 {
			fileTreeContentHeight = 1
		}
		fileTreeBorderColor := surface
		if m.focused == types.PaneFileTree {
			fileTreeBorderColor = green
		}
		fileTreeBorder := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(fileTreeBorderColor)

		fileTreePane := fileTreeBorder.
			Width(fileTreeWidth - 2).
			Height(fileTreeContentHeight).
			Render(m.fileTree.View())

		leftPane := lipgloss.JoinVertical(lipgloss.Left, commitInputView, fileTreePane)
		content = lipgloss.JoinHorizontal(lipgloss.Top, leftPane, diffViewPane)
	}

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

	m.statusBar.SetWidth(frameWidth - 2)
	statusBar := m.statusBar.View()

	innerContent := lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		content,
		statusBar,
	)

	outerFrame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(surface).
		Width(frameWidth).
		Height(frameHeight)

	return m.newView(outerFrame.Render(innerContent))
}
