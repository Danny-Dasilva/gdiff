package diffview

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Danny-Dasilva/gdiff/internal/types"
	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

func newTestKeyMap() types.KeyMap {
	return types.KeyMap{
		Up: key.NewBinding(key.WithKeys("k")),
		Down: key.NewBinding(key.WithKeys("j")),
		Left: key.NewBinding(key.WithKeys("h")),
		Right: key.NewBinding(key.WithKeys("l")),
		VisualMode: key.NewBinding(key.WithKeys("v")),
		Escape: key.NewBinding(key.WithKeys("esc")),
	}
}

func TestCharacterSelection(t *testing.T) {
	m := New(newTestKeyMap())
	m.SetFocused(true)

	// Set up a diff with an added line
	diffs := []diff.FileDiff{
		{
			OldPath: "test.go",
			NewPath: "test.go",
			Hunks: []diff.Hunk{
				{
					OldStart: 1,
					OldCount: 1,
					NewStart: 1,
					NewCount: 1,
					Lines: []diff.Line{
						{Type: diff.LineRemoved, Content: "old line content", OldNum: 1},
						{Type: diff.LineAdded, Content: "new line content here", NewNum: 1},
					},
				},
			},
		},
	}
	m.SetDiff("test.go", diffs)
	m.SetSize(80, 24)

	// Move cursor to the added line (line 2 in flattened view: header at 0, removed at 1, added at 2)
	m.moveCursor(2)

	t.Run("character mode activates with v", func(t *testing.T) {
		// Enter visual mode
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})

		if !m.visualMode {
			t.Error("expected visual mode to be enabled")
		}
		if !m.charMode {
			t.Error("expected character mode to be enabled on added/removed line")
		}
	})

	t.Run("h/l moves character cursor in char mode", func(t *testing.T) {
		initialPos := m.charCursor

		// Move right
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if m.charCursor != initialPos+1 {
			t.Errorf("charCursor should advance, got %d want %d", m.charCursor, initialPos+1)
		}

		// Move left
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if m.charCursor != initialPos {
			t.Errorf("charCursor should go back, got %d want %d", m.charCursor, initialPos)
		}
	})

	t.Run("character selection range updates", func(t *testing.T) {
		// Reset to start
		m.charStart = 0
		m.charCursor = 0

		// Move cursor to position 5
		for i := 0; i < 5; i++ {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		}

		if m.charStart != 0 {
			t.Errorf("charStart should stay at 0, got %d", m.charStart)
		}
		if m.charCursor != 5 {
			t.Errorf("charCursor should be 5, got %d", m.charCursor)
		}
	})

	t.Run("escape exits character mode", func(t *testing.T) {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})

		if m.charMode {
			t.Error("character mode should be disabled after escape")
		}
		if m.visualMode {
			t.Error("visual mode should be disabled after escape")
		}
	})
}

func TestCharacterSelectionBounds(t *testing.T) {
	m := New(newTestKeyMap())
	m.SetFocused(true)

	diffs := []diff.FileDiff{
		{
			OldPath: "test.go",
			NewPath: "test.go",
			Hunks: []diff.Hunk{
				{
					Lines: []diff.Line{
						{Type: diff.LineAdded, Content: "short", NewNum: 1},
					},
				},
			},
		},
	}
	m.SetDiff("test.go", diffs)
	m.SetSize(80, 24)

	// Move to added line and enter visual mode
	m.moveCursor(1)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})

	t.Run("charCursor cannot go below 0", func(t *testing.T) {
		// Try to move left past start
		for i := 0; i < 10; i++ {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		}
		if m.charCursor < 0 {
			t.Errorf("charCursor should not be negative, got %d", m.charCursor)
		}
	})

	t.Run("charCursor cannot exceed line length", func(t *testing.T) {
		lineLen := 5 // "short"
		// Try to move right past end
		for i := 0; i < 20; i++ {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		}
		if m.charCursor > lineLen {
			t.Errorf("charCursor should not exceed line length %d, got %d", lineLen, m.charCursor)
		}
	})
}

func TestGetCharacterSelection(t *testing.T) {
	m := New(newTestKeyMap())
	m.SetFocused(true)

	content := "hello world"
	diffs := []diff.FileDiff{
		{
			OldPath: "test.go",
			NewPath: "test.go",
			Hunks: []diff.Hunk{
				{
					Lines: []diff.Line{
						{Type: diff.LineAdded, Content: content, NewNum: 1},
					},
				},
			},
		},
	}
	m.SetDiff("test.go", diffs)
	m.SetSize(80, 24)
	m.moveCursor(1)

	// Enter visual mode and select characters 0-5
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	for i := 0; i < 5; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	}

	sel := m.GetCharacterSelection()

	if sel == nil {
		t.Fatal("expected non-nil character selection")
	}
	if sel.LineIndex != 1 {
		t.Errorf("LineIndex = %d, want 1", sel.LineIndex)
	}
	if sel.Start != 0 {
		t.Errorf("Start = %d, want 0", sel.Start)
	}
	if sel.End != 5 {
		t.Errorf("End = %d, want 5", sel.End)
	}
}

func TestCharacterModeOnlyForChangedLines(t *testing.T) {
	m := New(newTestKeyMap())
	m.SetFocused(true)

	diffs := []diff.FileDiff{
		{
			OldPath: "test.go",
			NewPath: "test.go",
			Hunks: []diff.Hunk{
				{
					Lines: []diff.Line{
						{Type: diff.LineContext, Content: "context line", OldNum: 1, NewNum: 1},
						{Type: diff.LineAdded, Content: "added line", NewNum: 2},
					},
				},
			},
		},
	}
	m.SetDiff("test.go", diffs)
	m.SetSize(80, 24)

	// Move to context line (index 1, since header is 0)
	m.moveCursor(1)

	// Enter visual mode on context line
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})

	if m.charMode {
		t.Error("character mode should NOT activate on context lines")
	}

	// But visual mode should still work
	if !m.visualMode {
		t.Error("visual mode should still work on context lines")
	}
}
