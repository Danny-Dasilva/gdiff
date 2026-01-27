package types

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for the application
type KeyMap struct {
	// Navigation
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Top       key.Binding
	Bottom    key.Binding
	HalfUp    key.Binding
	HalfDown  key.Binding
	NextHunk  key.Binding
	PrevHunk  key.Binding
	NextChange key.Binding
	PrevChange key.Binding

	// Pane switching
	SwitchPane key.Binding

	// Selection
	VisualMode   key.Binding
	VisualLine   key.Binding
	SelectHunk   key.Binding

	// Staging
	StageFile    key.Binding
	UnstageFile  key.Binding
	StageItem    key.Binding
	UnstageItem  key.Binding
	StageHunk    key.Binding
	UnstageHunk  key.Binding
	RevertItem   key.Binding

	// View toggle
	ToggleStagedView key.Binding

	// Commit/Push
	Commit      key.Binding
	CommitAmend key.Binding
	Push        key.Binding
	ForcePush   key.Binding

	// Search
	Search     key.Binding
	SearchNext key.Binding
	SearchPrev key.Binding

	// General
	Help   key.Binding
	Quit   key.Binding
	Enter  key.Binding
	Escape key.Binding

	// Editor
	OpenEditor key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/\u2191", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/\u2193", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/\u2190", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/\u2192", "right"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("gg", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		HalfUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("^u", "half page up"),
		),
		HalfDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("^d", "half page down"),
		),
		NextHunk: key.NewBinding(
			key.WithKeys("}"),
			key.WithHelp("}", "next hunk"),
		),
		PrevHunk: key.NewBinding(
			key.WithKeys("{"),
			key.WithHelp("{", "prev hunk"),
		),
		NextChange: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]c", "next change"),
		),
		PrevChange: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[c", "prev change"),
		),

		// Pane switching
		SwitchPane: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),

		// Selection
		VisualMode: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "visual mode"),
		),
		VisualLine: key.NewBinding(
			key.WithKeys("V"),
			key.WithHelp("V", "visual line"),
		),
		SelectHunk: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select hunk"),
		),

		// Staging (using plan's keybindings)
		StageFile: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "stage file"),
		),
		UnstageFile: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "unstage file"),
		),
		StageItem: key.NewBinding(
			key.WithKeys("s", " "),
			key.WithHelp("s/space", "stage selection"),
		),
		UnstageItem: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "unstage selection"),
		),
		RevertItem: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "revert (confirm)"),
		),
		StageHunk: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "stage hunk"),
		),
		UnstageHunk: key.NewBinding(
			key.WithKeys("U"),
			key.WithHelp("U", "unstage hunk"),
		),
		ToggleStagedView: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle staged view"),
		),

		// Commit/Push
		Commit: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "commit"),
		),
		CommitAmend: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "amend commit"),
		),
		Push: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "push"),
		),
		ForcePush: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "force push"),
		),

		// Search
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		SearchNext: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next match"),
		),
		SearchPrev: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev match"),
		),

		// General
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),

		// Editor
		OpenEditor: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("^e", "open in editor"),
		),
	}
}
