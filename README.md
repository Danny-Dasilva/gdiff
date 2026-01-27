# gdiff

A powerful Git diff TUI with Vim-style navigation and granular staging control.

![Demo](demo.gif)

## Features

- **VS Code-style interface** - Familiar layout with staged/unstaged sections and inline commit input
- **Two-pane layout** - File tree with sections on the left, diff view on the right
- **Vim-style navigation** - j/k, gg/G, Ctrl+d/u, and more
- **Granular staging** - Stage files, hunks, lines, or even individual characters
- **Visual selection** - Select exactly what you want to stage
- **Inline commit** - Type commit message directly in the UI (press `i`)
- **Commit integration** - Built-in commit dialog with $EDITOR support
- **Push support** - Push and force-push from within the TUI
- **Collapsible sections** - Expand/collapse staged and unstaged changes
- **File icons** - Language-specific icons (requires Nerd Font)
- **Async operations** - Non-blocking with spinners for long operations
- **Diff caching** - Fast navigation with intelligent cache invalidation

## Installation

### From source

```bash
go install github.com/Danny-Dasilva/gdiff/cmd/gdiff@latest
```

### Build locally

```bash
git clone https://github.com/Danny-Dasilva/gdiff
cd gdiff
go build -o gdiff ./cmd/gdiff
```

## Usage

```bash
# Run in any git repository
gdiff

# Or specify a path
gdiff /path/to/repo
```

## Keybindings

### Navigation

![Navigation Demo](demo-navigation.gif)

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `h` / `l` | Move left / right (in visual mode) |
| `gg` | Go to top |
| `G` | Go to bottom |
| `Ctrl+d` | Half page down |
| `Ctrl+u` | Half page up |
| `}` | Jump to next hunk |
| `{` | Jump to previous hunk |
| `]c` | Jump to next change |
| `[c` | Jump to previous change |
| `Tab` | Switch between file tree and diff pane |

### Staging

![Staging Demo](demo-staging.gif)

| Key | Action |
|-----|--------|
| `a` | Stage file (in file tree) |
| `A` | Unstage file |
| `s` / `Space` | Stage selection |
| `u` | Unstage selection |
| `S` | Stage entire hunk |
| `U` | Unstage entire hunk |
| `x` | Revert changes (with confirmation) |

### Visual Selection

![Character-Level Diff Demo](demo-character-diff.gif)

| Key | Action |
|-----|--------|
| `v` | Enter character visual mode |
| `V` | Enter line visual mode |
| `h` / `l` | Expand selection left / right |
| `j` / `k` | Expand selection up / down |
| `Escape` | Exit visual mode |

### Commit & Push

![Commit Demo](demo-commit.gif)

| Key | Action |
|-----|--------|
| `i` | Focus inline commit input (press Enter to commit, Esc to cancel) |
| `c` | Open commit dialog |
| `C` | Amend last commit |
| `Ctrl+e` | Open $EDITOR for commit message |
| `p` | Push to remote |
| `P` | Force push to remote |

### View

| Key | Action |
|-----|--------|
| `t` | Toggle between staged/unstaged view |
| `?` | Show help |
| `q` | Quit |

## Workflow Examples

### Stage specific lines from a file

1. Navigate to the file in the file tree
2. Press `Tab` to switch to the diff view
3. Press `v` to enter visual mode
4. Use `j`/`k` to select lines
5. Press `s` to stage the selection

### Stage specific characters

1. Navigate to the line you want to partially stage
2. Press `v` to enter visual mode
3. Use `h`/`l` to select specific characters
4. Press `s` to stage just those characters

### Quick file staging

1. Use `j`/`k` to navigate the file tree
2. Press `a` to stage entire files
3. Press `c` to commit when ready

### Review staged changes

1. Press `t` to toggle to staged view
2. Review what will be committed
3. Press `t` again to return to unstaged view

## Architecture

Built with the [Charmbracelet](https://charm.sh/) stack:

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework (Elm architecture)
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - TUI components (viewport, textinput)
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Style definitions

### Project Structure

```
cmd/gdiff/          # Entry point
internal/
  app/              # Main application model
  config/           # Configuration
  git/              # Git operations (status, diff, staging)
  types/            # Shared types and keybindings
  ui/
    filetree/       # File tree component
    diffview/       # Diff view component
    statusbar/      # Status bar component
    commit/         # Commit modal component
    spinner/        # Loading spinner component
pkg/
  diff/             # Diff parsing and highlighting
```

## Technical Details

- Uses `git diff --histogram` for better code diff grouping
- Character-level staging via `git apply --cached` with custom patches
- LCS algorithm for character-level change detection within lines
- Async diff loading with context cancellation for responsiveness
- Diff caching with automatic invalidation on staging operations

## Requirements

- Go 1.21+
- Git 2.0+

## License

MIT
