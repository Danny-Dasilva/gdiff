package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danny/gdiff/internal/app"
	"github.com/danny/gdiff/internal/git"
)

func main() {
	// Check if we're in a git repository
	ctx := context.Background()
	if !git.IsGitRepo(ctx) {
		fmt.Fprintln(os.Stderr, "Error: not a git repository")
		os.Exit(1)
	}

	// Create and run the application
	p := tea.NewProgram(
		app.New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
