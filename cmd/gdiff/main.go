package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Danny-Dasilva/gdiff/internal/app"
	"github.com/Danny-Dasilva/gdiff/internal/git"
)

func main() {
	colorblind := flag.Bool("colorblind", false, "use blue/orange colors instead of red/green for colorblind accessibility")
	flag.Parse()

	// Check if we're in a git repository
	ctx := context.Background()
	if !git.IsGitRepo(ctx) {
		fmt.Fprintln(os.Stderr, "Error: not a git repository")
		os.Exit(1)
	}

	// Create and run the application
	p := tea.NewProgram(
		app.New(*colorblind),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
