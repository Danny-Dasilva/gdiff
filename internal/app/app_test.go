package app

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Danny-Dasilva/gdiff/internal/types"
)

// TestCancelFuncFieldExists verifies the Model has a cancelDiffLoad field
func TestCancelFuncFieldExists(t *testing.T) {
	m := New()
	// The cancel func should be nil initially
	if m.cancelDiffLoad != nil {
		t.Error("cancelDiffLoad should be nil on new model")
	}
}

// TestNewDiffLoadCancelsPrevious verifies that starting a new diff load
// cancels any pending previous load
func TestNewDiffLoadCancelsPrevious(t *testing.T) {
	m := New()

	// Set up a cancel function that we can track
	var cancelled atomic.Bool
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelDiffLoad = cancel

	// Simulate starting a new diff load by calling loadDiff
	// This should cancel the previous context
	_ = m.loadDiff("new-file.go", false)

	// The old cancel should have been called
	select {
	case <-ctx.Done():
		cancelled.Store(true)
	case <-time.After(100 * time.Millisecond):
		// Give it a moment
	}

	if !cancelled.Load() {
		t.Error("Previous diff load was not cancelled when new load started")
	}
}

// TestLoadDiffRespectsContext verifies loadDiff respects context cancellation
func TestLoadDiffRespectsContext(t *testing.T) {
	m := New()

	// Start a diff load
	cmd := m.loadDiff("test.go", false)

	// The command should exist
	if cmd == nil {
		t.Fatal("loadDiff should return a command")
	}

	// After loading, cancelDiffLoad should be set
	if m.cancelDiffLoad == nil {
		t.Error("cancelDiffLoad should be set after loadDiff")
	}
}

// TestFileNavigationCancelsPendingLoad verifies that navigating to a different
// file cancels the pending diff load
func TestFileNavigationCancelsPendingLoad(t *testing.T) {
	m := New()
	m.width = 100
	m.height = 50
	m.updateLayout()

	// Track cancellation
	var cancelCount atomic.Int32
	var mu sync.Mutex

	// Mock the cancel function
	mu.Lock()
	ctx1, cancel1 := context.WithCancel(context.Background())
	m.cancelDiffLoad = func() {
		cancelCount.Add(1)
		cancel1()
	}
	mu.Unlock()

	// Simulate file selection message (navigating to a different file)
	msg := types.FileSelectedMsg{Path: "different-file.go", Staged: false}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// The previous cancel should have been called
	select {
	case <-ctx1.Done():
		// Good, context was cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should have been cancelled on file navigation")
	}

	if cancelCount.Load() < 1 {
		t.Error("Cancel function should have been called on file navigation")
	}
}

// TestCachedDiffDoesNotCancelPrevious verifies that a cached diff hit
// does not unnecessarily cancel (since it returns immediately)
func TestCachedDiffStillCancelsPrevious(t *testing.T) {
	m := New()

	// Pre-populate cache
	m.diffCache["cached-file.go"] = nil

	// Set up tracking for cancellation
	var cancelled atomic.Bool
	_, cancel := context.WithCancel(context.Background())
	m.cancelDiffLoad = func() {
		cancelled.Store(true)
		cancel()
	}

	// Load a cached file
	_ = m.loadDiff("cached-file.go", false)

	// Even for cached loads, we should cancel any pending async load
	// to maintain consistent state
	if !cancelled.Load() {
		t.Error("Should still cancel previous load even when returning cached result")
	}
}

// TestDiffLoadedMsgForStalePathIgnored verifies that if a DiffLoadedMsg
// arrives for a file that's no longer the current file, it's handled gracefully
func TestDiffLoadedMsgUpdatesCurrentFile(t *testing.T) {
	m := New()
	m.width = 100
	m.height = 50
	m.updateLayout()

	// Receive a DiffLoadedMsg
	msg := types.DiffLoadedMsg{Path: "loaded-file.go", Diffs: nil, Err: nil}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.currentFile != "loaded-file.go" {
		t.Errorf("currentFile should be updated to 'loaded-file.go', got '%s'", m.currentFile)
	}
}

// TestContextPassedToGetFileDiff verifies that the context is properly
// passed through to the git command (integration-style test)
func TestLoadDiffCreatesNewCancelFunc(t *testing.T) {
	m := New()

	// Initial state
	if m.cancelDiffLoad != nil {
		t.Error("cancelDiffLoad should be nil initially")
	}

	// After loadDiff for non-cached file, should have cancel func
	_ = m.loadDiff("uncached-file.go", false)

	if m.cancelDiffLoad == nil {
		t.Error("cancelDiffLoad should be set after loadDiff")
	}
}
