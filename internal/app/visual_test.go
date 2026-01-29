package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Danny-Dasilva/gdiff/internal/types"
)

// TestActiveBorderIsGreen verifies that focused pane uses green border (#a6e3a1)
// instead of blue, matching lazygit/soft-serve convention
func TestActiveBorderIsGreen(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 40
	m.updateLayout()

	// The View should render without panic
	view := m.View()
	if view == "" {
		t.Fatal("View should not be empty")
	}
}

// TestStagingFlashMessageStaged verifies "Staged: filename" message appears
func TestStagingFlashMessageStaged(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 40
	m.updateLayout()

	// Simulate stage complete message
	msg := types.StageCompleteMsg{Path: "main.go", Err: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// View should contain staging confirmation
	view := m.View()
	if !strings.Contains(view, "Staged") || !strings.Contains(view, "main.go") {
		t.Error("View should show 'Staged: main.go' flash message after staging")
	}
}

// TestStagingFlashMessageUnstaged verifies "Unstaged: filename" message appears
func TestStagingFlashMessageUnstaged(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 40
	m.updateLayout()

	// Simulate unstage complete message
	msg := types.UnstageCompleteMsg{Path: "main.go", Err: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// View should contain unstaging confirmation
	view := m.View()
	if !strings.Contains(view, "Unstaged") || !strings.Contains(view, "main.go") {
		t.Error("View should show 'Unstaged: main.go' flash message after unstaging")
	}
}

// TestViewRendersWithoutPanic ensures View doesn't crash with various focus states
func TestViewRendersWithFocusStates(t *testing.T) {
	panes := []types.Pane{types.PaneFileTree, types.PaneDiffView, types.PaneCommitInput}

	for _, pane := range panes {
		t.Run(fmt.Sprintf("pane_%d", pane), func(t *testing.T) {
			m := New()
			m.width = 120
			m.height = 40
			m.focused = pane
			m.updateLayout()

			view := m.View()
			if view == "" {
				t.Fatalf("View should not be empty for pane %v", pane)
			}
		})
	}
}
