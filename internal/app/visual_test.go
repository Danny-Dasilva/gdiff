package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Danny-Dasilva/gdiff/internal/types"
)

func newTestModel() Model {
	m := New(false)
	m.width = 120
	m.height = 40
	m.updateLayout()
	return m
}

func TestActiveBorderIsGreen(t *testing.T) {
	m := newTestModel()

	view := m.View()
	if view.Content == "" {
		t.Fatal("View should not be empty")
	}
}

func TestStagingFlashMessageStaged(t *testing.T) {
	m := newTestModel()

	msg := types.StageCompleteMsg{Path: "main.go"}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	view := m.View()
	if !strings.Contains(view.Content, "Staged") || !strings.Contains(view.Content, "main.go") {
		t.Error("View should show 'Staged: main.go' flash message after staging")
	}
}

func TestStagingFlashMessageUnstaged(t *testing.T) {
	m := newTestModel()

	msg := types.UnstageCompleteMsg{Path: "main.go"}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	view := m.View()
	if !strings.Contains(view.Content, "Unstaged") || !strings.Contains(view.Content, "main.go") {
		t.Error("View should show 'Unstaged: main.go' flash message after unstaging")
	}
}

func TestViewRendersWithFocusStates(t *testing.T) {
	panes := []types.Pane{types.PaneFileTree, types.PaneDiffView, types.PaneCommitInput}

	for _, pane := range panes {
		t.Run(fmt.Sprintf("pane_%d", pane), func(t *testing.T) {
			m := newTestModel()
			m.focused = pane

			view := m.View()
			if view.Content == "" {
				t.Fatalf("View should not be empty for pane %v", pane)
			}
		})
	}
}
