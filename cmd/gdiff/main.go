package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/Danny-Dasilva/gdiff/internal/app"
	"github.com/Danny-Dasilva/gdiff/internal/git"
)

func main() {
	colorblind := flag.Bool("colorblind", false, "use blue/orange colors instead of red/green for colorblind accessibility")
	flag.Parse()

	if !git.IsGitRepo(context.Background()) {
		fmt.Fprintln(os.Stderr, "Error: not a git repository")
		os.Exit(1)
	}

	if _, err := tea.NewProgram(app.New(*colorblind)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
