package helpoverlay

import (
	"regexp"
	"strings"
	"testing"
)

// stripAnsi removes ANSI escape sequences for test assertions
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func TestNewReturnsHiddenModel(t *testing.T) {
	m := New()
	if m.Visible() {
		t.Error("help overlay should be hidden by default")
	}
}

func TestToggleVisibility(t *testing.T) {
	m := New()

	m.Toggle()
	if !m.Visible() {
		t.Error("help overlay should be visible after first toggle")
	}

	m.Toggle()
	if m.Visible() {
		t.Error("help overlay should be hidden after second toggle")
	}
}

func TestHide(t *testing.T) {
	m := New()
	m.Toggle() // show
	m.Hide()
	if m.Visible() {
		t.Error("help overlay should be hidden after Hide()")
	}
}

func TestSetSize(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Errorf("SetSize failed: got width=%d height=%d, want 120x40", m.width, m.height)
	}
}

func TestViewWhenHidden(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	view := m.View()
	if view != "" {
		t.Error("View() should return empty string when hidden")
	}
}

func visibleView() string {
	m := New()
	m.SetSize(120, 40)
	m.Toggle()
	return m.View()
}

func TestViewContainsKeybindingSections(t *testing.T) {
	// Strip ANSI since v2 lipgloss wraps each char in underline styles
	plain := stripAnsi(visibleView())
	for _, section := range []string{"Navigation", "Staging", "View", "Commit", "Push"} {
		if !strings.Contains(plain, section) {
			t.Errorf("View() should contain section header %q", section)
		}
	}
}

func TestViewContainsKeyBindings(t *testing.T) {
	view := visibleView()
	for _, binding := range []string{"j/k", "Space", "quit", "help", "push"} {
		if !strings.Contains(view, binding) {
			t.Errorf("View() should contain key binding text %q", binding)
		}
	}
}

func TestViewContainsCloseHint(t *testing.T) {
	view := visibleView()
	if !strings.Contains(view, "?") && !strings.Contains(view, "Esc") {
		t.Error("View() should contain close hint mentioning ? or Esc")
	}
}

func TestViewContainsKeybindingsTitle(t *testing.T) {
	if !strings.Contains(visibleView(), "Keybindings") {
		t.Error("View() should contain 'Keybindings' title")
	}
}
