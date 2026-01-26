package types

import (
	"github.com/danny/gdiff/pkg/diff"
)

// StatusLoadedMsg is sent when git status is loaded
type StatusLoadedMsg struct {
	Files []diff.FileEntry
	Err   error
}

// DiffLoadedMsg is sent when a file's diff is loaded
type DiffLoadedMsg struct {
	Path  string
	Diffs []diff.FileDiff
	Err   error
}

// StageCompleteMsg is sent when a staging operation completes
type StageCompleteMsg struct {
	Path string
	Err  error
}

// UnstageCompleteMsg is sent when an unstaging operation completes
type UnstageCompleteMsg struct {
	Path string
	Err  error
}

// CommitCompleteMsg is sent when a commit completes
type CommitCompleteMsg struct {
	Hash string
	Err  error
}

// PushCompleteMsg is sent when a push completes
type PushCompleteMsg struct {
	Err error
}

// ErrorMsg represents a general error
type ErrorMsg struct {
	Err error
}

// FileSelectedMsg is sent when a file is selected in the file tree
type FileSelectedMsg struct {
	Path   string
	Staged bool
}

// FocusChangedMsg is sent when focus changes between panes
type FocusChangedMsg struct {
	Pane Pane
}

// Pane identifies UI panes
type Pane int

const (
	PaneFileTree Pane = iota
	PaneDiffView
)
