package diffview

import (
	"strings"
	"testing"

	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

// TestDualLineNumbers verifies that renderLine produces dual line numbers
// Format: "  42  56 | code here"
func TestDualLineNumbers(t *testing.T) {
	m := New(newTestKeyMap(), false)
	m.SetFocused(true)
	m.SetSize(120, 40)

	t.Run("added line shows blank old num and new num", func(t *testing.T) {
		line := diff.Line{Type: diff.LineAdded, Content: "new code", NewNum: 56}
		rendered := m.renderLine(line, 0)
		// Should contain the new line number but blank old
		if !strings.Contains(rendered, "56") {
			t.Errorf("added line should contain new line number 56, got: %s", rendered)
		}
	})

	t.Run("removed line shows old num and blank new num", func(t *testing.T) {
		line := diff.Line{Type: diff.LineRemoved, Content: "old code", OldNum: 42}
		rendered := m.renderLine(line, 0)
		// Should contain the old line number
		if !strings.Contains(rendered, "42") {
			t.Errorf("removed line should contain old line number 42, got: %s", rendered)
		}
	})

	t.Run("context line shows both line numbers", func(t *testing.T) {
		line := diff.Line{Type: diff.LineContext, Content: "unchanged", OldNum: 10, NewNum: 15}
		rendered := m.renderLine(line, 0)
		if !strings.Contains(rendered, "10") {
			t.Errorf("context line should contain old line number 10, got: %s", rendered)
		}
		if !strings.Contains(rendered, "15") {
			t.Errorf("context line should contain new line number 15, got: %s", rendered)
		}
	})

	t.Run("hunk header shows @@ markers", func(t *testing.T) {
		line := diff.Line{Type: diff.LineHunkHeader, Content: "@@ -1,3 +1,4 @@"}
		rendered := m.renderLine(line, 0)
		if !strings.Contains(rendered, "@@") {
			t.Errorf("hunk header should contain @@ markers, got: %s", rendered)
		}
	})
}

// TestDiffRenderLineMarkers verifies the +/- markers in rendered lines
func TestDiffRenderLineMarkers(t *testing.T) {
	m := New(newTestKeyMap(), false)
	m.SetFocused(true)
	m.SetSize(120, 40)

	t.Run("added line has + marker", func(t *testing.T) {
		line := diff.Line{Type: diff.LineAdded, Content: "added", NewNum: 1}
		rendered := m.renderLine(line, 0)
		if !strings.Contains(rendered, "+") {
			t.Errorf("added line should have + marker, got: %s", rendered)
		}
	})

	t.Run("removed line has - marker", func(t *testing.T) {
		line := diff.Line{Type: diff.LineRemoved, Content: "removed", OldNum: 1}
		rendered := m.renderLine(line, 0)
		if !strings.Contains(rendered, "-") {
			t.Errorf("removed line should have - marker, got: %s", rendered)
		}
	})
}

// TestDiffSeparatorPresent verifies the | separator between line numbers and code
func TestDiffSeparatorPresent(t *testing.T) {
	m := New(newTestKeyMap(), false)
	m.SetFocused(true)
	m.SetSize(120, 40)

	line := diff.Line{Type: diff.LineContext, Content: "code here", OldNum: 1, NewNum: 1}
	rendered := m.renderLine(line, 0)
	// The separator character should be present (rendered with styling)
	if !strings.Contains(rendered, "|") && !strings.Contains(rendered, "\u2502") {
		t.Errorf("rendered line should contain separator, got: %s", rendered)
	}
}
