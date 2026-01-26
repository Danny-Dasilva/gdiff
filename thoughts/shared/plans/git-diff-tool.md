# Implementation Plan: Go Git Diff Tool with Charmbracelet

Generated: 2026-01-26

## Goal

Build a fast terminal-based Git diff viewer and staging tool in Go using the Charmbracelet ecosystem (Bubble Tea, Bubbles, Lip Gloss, optional Glow/Gum). It must support file-, line-, and character-level diffs plus vim-style motions for staging/unstaging, committing, and pushing changes.

## Non-Goals

- Full Git porcelain replacement (branch mgmt, rebase, etc.) beyond commit/push scope.
- IDE-like refactors or merge tools.
- Replacing `git diff` output formatting tools for non-interactive use (we can integrate with them later).

## Research Summary (Charmbracelet Stack)

### Bubble Tea (framework)
- Go TUI framework based on The Elm Architecture with `Init`, `Update`, `View` methods and a model-driven loop; suited for both simple and complex terminal apps. It includes a framerate-based renderer, mouse support, and focus reporting, which is relevant for performance and responsiveness in large diffs. [R1]

### Bubbles (components)
- UI components (list, viewport, table, text input, spinner, paginator, file picker, etc.) used in production in Glow. This provides off-the-shelf components for list navigation, diff viewport, and modal input. [R2]

### Lip Gloss (styling)
- Declarative terminal styling library with adaptive colors and layout features (padding, borders, widths) designed for TUIs. This supports a consistent diff theme and accessibility-friendly styles. [R3]

### Glow (markdown rendering)
- Terminal markdown reader with TUI and CLI modes; discovers markdown in directories or git repos and renders with a high-performance pager. Useful for in-app help, changelog, or commit template preview. [R4]

### Gum (shell UX)
- A CLI tool for building rich prompts in shell scripts using Bubbles and Lip Gloss primitives. Useful as a fallback for non-TUI flows (e.g., scripted commit prompt or quick staging helper). [R5]

### Crush (conventions)
- Charmbracelet CLI with a documented local/global config discovery order and a simple JSON config model. We can mirror this for a predictable config system. [R6]

## Research Summary (Similar Tools and Gaps)

### Lazygit
- Supports staging individual lines and ranges with keybindings (line-level staging is mature). [R7]
- Gap: Does not provide explicit character-level staging or a dedicated character-diff selection workflow.

### GitUI
- Terminal UI for git with staging/unstaging of files, hunks, and lines, plus commit and push/fetch. [R8]
- Gap: Character-level selection and staging is not a first-class interaction model.

### Tig
- Ncurses git browser that can assist staging at chunk (hunk) level and act as a pager for git output. [R9]
- Gap: Line-level and character-level staging not emphasized; UI is more browser/pager oriented than editing diffs.

### Delta / diff-so-fancy
- Delta is a syntax-highlighting pager with within-line highlights, side-by-side view, and navigation across files; diff-so-fancy improves readability of unified diffs. These are excellent renderers but are not interactive staging tools. [R10] [R11]
- Gap: No interactive staging/unstaging or commit/push workflows.

### Key Opportunities
- Combine high-performance rendering with fine-grained staging (character-level) and vim-first navigation.
- Provide a faster, more focused staging experience than full Git UIs while offering deeper diff granularity than pagers.

## Feature Proposal

### Core Features
- File tree with status (modified, added, deleted, renamed, staged/unstaged).
- Diff view with:
  - File-level summaries and counts.
  - Hunk-level and line-level selection.
  - Character-level diff highlighting.
- Staging actions:
  - Stage/unstage file, hunk, line, or selected characters.
  - Revert changes at file/hunk/line/char level (safe prompts).
- Commit flow:
  - Commit message editor (inline or $EDITOR).
  - Optional staging checks (e.g., warn on empty staged set).
- Push flow:
  - Push to upstream with status display and error handling.
  - If upstream not set, prompt for remote/branch.
- Vim-style navigation and selection.
- Keyboard-only UX with contextual help.

### Differentiators
- Character-level staging for partial line edits (unique vs most TUI Git tools).
- High-performance diff virtualization for huge repos.
- Built for speed: minimal background work, aggressive caching, and async git operations.

## UX Layout and Interaction

### Layout
- Left panel: file list with status badges and staged/unstaged toggles.
- Right panel: diff view with hunks and inline character highlights.
- Bottom: status bar with key hints, mode, and git operation feedback.
- Modal overlays: commit message editor, push confirmation, error dialog.

### Vim Motions and Keymap (draft)

Navigation:
- `j/k` move line; `h/l` move char; `gg/G` top/bottom; `}` / `{` next/prev hunk; `]c` / `[c` next/prev change.

Selection:
- `v` toggle visual selection; `V` linewise selection; `ctrl+v` block selection (optional for columns).

Staging:
- `s` stage selection; `u` unstage selection.
- `S` stage hunk; `U` unstage hunk.
- `a` stage file; `A` unstage file.
- `x` revert selection (with confirm).

Commit/Push:
- `c` open commit modal; `C` amend last commit (optional).
- `p` push; `P` push --force-with-lease (optional, confirm).

Help:
- `?` help overlay; `:` command palette (optional).

## Diff Engine and Staging Design

### Git Data Acquisition
- `git status --porcelain=v2 -z` for file list and status.
- `git diff --no-color --unified=0 --histogram` for focused file (fast, minimal context).
- `git diff --cached` for staged view toggles.
- `git diff --numstat` for lightweight stats when diff is not yet loaded.

### Parsing
- Parse unified diff into structures:
  - `FileDiff` -> `Hunk` -> `Line` -> `Token` (for char diffs).
- Character-level highlighting computed only for lines in view (lazy).

### Character-Level Diff
- For each modified line pair, compute a minimal edit script (Myers or similar) to derive character insertions/deletions.
- Render edits with fine-grained highlights and selection handles.

### Staging Mechanics
- File staging: `git add <path>` / `git reset <path>`.
- Hunk/line/char staging:
  1) Build a synthetic patch from the selection with exact line ranges.
  2) Apply with `git apply --cached` (and `--unidiff-zero` if needed).
  3) Validate by reloading staged diff.
- Char-level staging uses line splitting: convert a single-line edit into a mini-hunk that deletes the old line and adds the selected new line with partial content. This enables granular staging without editing the working tree.
- Revert uses `git checkout -- <path>` or `git apply -R` with selection patch for finer granularity.

## Performance Strategy

- Lazy diff loading: only parse/render the focused file.
- Incremental diff updates: small changes only refresh impacted hunks.
- Virtualized viewport: render visible lines only; use Bubbles viewport in high-performance mode.
- Async git ops: background goroutines return typed messages; cancel previous diff loads on navigation.
- Fast-path status: `git status --porcelain=v2 -z` and `git diff --numstat` are fast and cacheable.
- Large files: warn and use simplified rendering or disable char diff past thresholds.

## Proposed Architecture

```
cmd/
  gdiff/
    main.go
internal/
  app/           # root model and routing
  git/           # git commands and parsing
  diff/          # diff parser + char diff engine
  ui/            # components, layout, styles
  config/        # config loading and defaults
```

### Core Models
- `AppModel`: focus management, routing, global keymap, window size.
- `FileListModel`: list of files + status + staged state.
- `DiffViewModel`: diff content, cursor, selection, viewport.
- `CommitModel`: commit text input, validation.
- `StatusBarModel`: key hints, errors, operation status.

## Config and CLI

- Config discovery inspired by Crush style:
  - `.gdiff.json`
  - `gdiff.json`
  - `$HOME/.config/gdiff/gdiff.json`
- CLI flags override config: `--repo`, `--cached`, `--theme`, `--no-color`, `--perf-mode`.

## Implementation Plan (Phases)

Phase 1: Repo and Status
- Detect repo, read status, render file list.
- Toggle staged/unstaged views.

Phase 2: Diff Parsing + View
- Parse diff for selected file and render hunks.
- Add navigation and selection at file/hunk/line level.

Phase 3: Staging Engine
- Stage/unstage file/hunk/line via patch application.
- Add revert at line/hunk levels.

Phase 4: Character-Level Diff
- Compute and render per-line char diffs.
- Enable character-level selection and staging.

Phase 5: Commit + Push
- Commit modal, optional $EDITOR support.
- Push action with upstream prompt.

Phase 6: Performance + UX Polish
- Diff caching, cancellation, and lazy rendering.
- Keymap help, error surfaces, and progress spinners.

Phase 7: Release
- Packaging (brew, deb/rpm), docs, example config.

## Testing Strategy

- Unit tests for diff parser and char diff engine.
- Snapshot tests for rendering (golden files).
- Integration tests against real git repos (temp fixtures).
- Performance tests on large diffs to validate latency and memory limits.

## Risks and Mitigations

- Char-level staging correctness: mitigate with strict patch validation and fallback to line-level.
- Large diffs and binary files: skip or summarize, warn user.
- Cross-platform git behavior: use consistent flags and validate on macOS/Linux/Windows.

## Open Questions

- Should we integrate external diff renderers (delta) for optional themes?
- How aggressive should char-level staging be on single-line edits (auto or manual)?
- What are acceptable thresholds for large file fallback modes?

## References

[R1] https://github.com/charmbracelet/bubbletea
[R2] https://github.com/charmbracelet/bubbles
[R3] https://github.com/charmbracelet/lipgloss
[R4] https://github.com/charmbracelet/glow
[R5] https://github.com/charmbracelet/gum
[R6] https://github.com/charmbracelet/crush
[R7] https://github.com/jesseduffield/lazygit
[R8] https://github.com/gitui-org/gitui
[R9] https://github.com/jonas/tig
[R10] https://dandavison.github.io/delta/features.html
[R11] https://github.com/so-fancy/diff-so-fancy
