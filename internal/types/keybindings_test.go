package types

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func makeKeyMsg(k string) tea.KeyMsg {
	// For special keys like ctrl+f, ctrl+b, enter
	switch k {
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
}

func TestDefaultKeyMap_FullPageScrollBindings(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("FullPageUp binding exists with ctrl+b", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("ctrl+b"), km.FullPageUp) {
			t.Error("FullPageUp should match ctrl+b")
		}
	})

	t.Run("FullPageDown binding exists with ctrl+f", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("ctrl+f"), km.FullPageDown) {
			t.Error("FullPageDown should match ctrl+f")
		}
	})
}

func TestDefaultKeyMap_RevertItemUsesD(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("RevertItem uses d key", func(t *testing.T) {
		if !key.Matches(makeKeyMsg("d"), km.RevertItem) {
			t.Error("RevertItem should match 'd' key")
		}
	})

	t.Run("RevertItem does not use x key", func(t *testing.T) {
		if key.Matches(makeKeyMsg("x"), km.RevertItem) {
			t.Error("RevertItem should NOT match 'x' key")
		}
	})
}

func TestDefaultKeyMap_SpaceToggle(t *testing.T) {
	km := DefaultKeyMap()

	t.Run("SpaceToggle binding exists with space", func(t *testing.T) {
		if !key.Matches(makeKeyMsg(" "), km.SpaceToggle) {
			t.Error("SpaceToggle should match space key")
		}
	})
}
