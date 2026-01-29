# Research Report: Git Diff Viewer UX Patterns
Generated: 2026-01-28

## Summary

Research across five major diff tools (lazygit, delta, difftastic, GitHub/GitLab web UIs, VS Code) reveals convergent UX patterns for diff viewing and staging. The best diff UIs combine: (1) character/word-level highlighting within changed lines, (2) multiple granularity levels for staging (file > hunk > line > character), (3) vim-style keyboard navigation with hunk-jump shortcuts, and (4) togglable unified vs side-by-side display. gdiff already implements many of these well; the gaps are primarily in display polish and advanced navigation.

## Questions Answered

### Q1: Diff Display Modes (Unified vs Side-by-Side)
**Answer:** All mature tools support both modes. Delta enables side-by-side with `-s` flag, with line numbers active by default in that mode. GitHub/GitLab both offer toggle between unified and split views, persisting the user's preference. Unified is preferred for single-line changes; side-by-side is better for large sequential changes across many lines. GitLab specifically notes: "Inline mode is often better for changes to single lines. Side-by-side mode is often better for changes affecting large numbers of sequential lines."
**Source:** [Delta GitHub](https://github.com/dandavison/delta), [GitLab Docs](https://docs.gitlab.com/user/project/merge_requests/changes/)
**Confidence:** High

### Q2: Line Numbering Styles
**Answer:** Delta shows dual line numbers (old/new) in side-by-side by default. GitHub shows line numbers for both sides in split view, clickable for linking. VS Code shows line numbers in both panes of its side-by-side diff editor. Best practice: show both old and new line numbers, make them clickable/linkable for reference.
**Source:** [Delta docs](https://dandavison.github.io/delta/introduction.html), [GitHub Docs](https://docs.github.com/articles/reviewing-proposed-changes-in-a-pull-request)
**Confidence:** High

### Q3: Change Highlighting (Character-Level)
**Answer:** Three tiers of highlighting exist in practice:
1. **Line-level** (basic): entire added/removed lines colored green/red
2. **Word-level** (delta): uses Levenshtein edit distance to highlight changed tokens within lines
3. **Character-level** (git --word-diff-regex=.): treats each character as a token

Delta's word-level approach using Levenshtein edit inference is the gold standard for readability. Git's built-in `diff-highlight` script does sub-word fragment highlighting by pairing matching lines. gdiff already uses LCS for character-level detection, which is competitive.
**Source:** [Delta GitHub](https://github.com/dandavison/delta), [Viget article](https://www.viget.com/articles/dress-up-your-git-diffs-with-word-level-highlights/)
**Confidence:** High

### Q4: Hunk Navigation Patterns
**Answer:** Consistent patterns across tools:
- **Lazygit**: scroll through hunks in staging view, `a` to select entire hunk, arrow keys between hunks
- **Delta**: `n`/`N` keys to jump between files in large diffs (--navigate mode)
- **VS Code**: `F7` / `Shift+F7` for next/previous change
- **Emacs diff-mode**: `M-n`/`M-p` for next/previous hunk
- **gdiff already has**: `}`/`{` for next/prev hunk, `]c`/`[c` for next/prev change (matches vim-diff convention)

gdiff's hunk navigation is already well-aligned with conventions.
**Source:** [Lazygit keybindings](https://github.com/jesseduffield/lazygit/blob/master/docs/keybindings/Keybindings_en.md), [VS Code docs](https://code.visualstudio.com/docs/sourcecontrol/overview)
**Confidence:** High

### Q5: Staging UX (What Makes It Intuitive)
**Answer:** Key insights from lazygit (the gold standard for TUI staging):
- **Space to stage** the current selection (line, hunk, or file depending on context)
- **`v` for visual range selection**, then space to stage the range
- **`a` to select entire hunk** while in staging view
- **Immediate visual feedback**: staged items move from "Unstaged" to "Staged" section
- **Toggle view**: ability to see staged vs unstaged view of the same file
- **Granularity ladder**: file → hunk → line → character, each accessible without mode switching

The critical UX insight from lazygit's philosophy: "To stage part of a file you need to use a command line program to step through each hunk... Lazygit simplifies this" by making it visual and direct.

VS Code adds: gutter icons for stage/revert actions on individual lines, which removes the need to enter a special mode.
**Source:** [Lazygit GitHub](https://github.com/jesseduffield/lazygit), [VS Code staging docs](https://code.visualstudio.com/docs/sourcecontrol/staging-commits)
**Confidence:** High

### Q6: Inline Actions (Stage Line, Revert, etc.)
**Answer:**
- **VS Code**: Gutter decorations with stage/revert buttons per line. Click to stage or revert individual lines directly.
- **GitHub/GitLab**: Comment icon appears on hover over line numbers. Multi-line selection for comments.
- **Lazygit**: Space stages current line, `x` discards. No separate "revert" UI; discarding is the equivalent.
- **gdiff already has**: `s`/`u` for stage/unstage, `x` for revert with confirmation, `S`/`U` for hunk-level.

gdiff's approach is solid. Consider adding: visual gutter indicators showing which lines are stageable.
**Source:** [VS Code docs](https://code.visualstudio.com/docs/sourcecontrol/staging-commits), [GitLab docs](https://docs.gitlab.com/user/project/merge_requests/changes/)
**Confidence:** High

### Q7: Context Lines (How Many? Collapsible?)
**Answer:**
- **Standard**: 3 context lines is the git default and universally expected minimum
- **Patch compatibility**: "patch typically needs at least two lines of context" for proper operation
- **GitHub**: Expandable context -- click to load more context lines above/below a hunk
- **GitLab**: Collapsible generated files by default; context expansion on demand
- **Best practice**: Start with 3 lines, allow expansion on demand. Never use 0 context in a viewer (usable only for inline diffs per Sublime Text issue)
- **Collapsibility**: GitHub and GitLab auto-collapse generated files, lock files, and minified files. Manual collapse per-file is standard.
**Source:** [GNU diff docs](https://www.gnu.org/software/diffutils/manual/html_node/diff-Options.html), [GitLab docs](https://docs.gitlab.com/user/project/merge_requests/changes/)
**Confidence:** High

## Feature Comparison Table

| Feature | lazygit | delta | difftastic | GitHub UI | VS Code | gdiff (current) |
|---------|---------|-------|------------|-----------|---------|-----------------|
| **Unified diff** | Yes | Yes | Yes | Yes | Yes | Yes |
| **Side-by-side** | No (uses pager) | Yes (`-s`) | Yes | Yes (toggle) | Yes | No |
| **Syntax highlighting** | Via delta pager | Yes (bat themes) | Yes (treesitter) | Yes | Yes | Partial |
| **Word-level highlight** | Via delta | Levenshtein | AST-aware | Yes | Yes | LCS character-level |
| **Line numbers** | Yes | Dual old/new | No | Dual | Dual | Yes |
| **Hunk navigation** | Scroll-based | `n`/`N` files | N/A | Scroll | `F7`/`Shift+F7` | `}`/`{` hunks |
| **File staging** | `space` | N/A (viewer) | N/A (viewer) | Web UI | `+` icon | `a` |
| **Hunk staging** | `a` (select all) | N/A | N/A | N/A | Gutter icon | `S` |
| **Line staging** | `space` on line | N/A | N/A | N/A | Gutter icon | `s` / `space` |
| **Char staging** | No | N/A | N/A | N/A | No | `v` + `h`/`l` + `s` |
| **Visual selection** | `v` range | N/A | N/A | Click-drag | Click-drag | `v`/`V` modes |
| **Revert/discard** | Yes | N/A | N/A | N/A | Gutter icon | `x` (with confirm) |
| **Context lines** | Git default (3) | Configurable | Full file | Expandable | Full file | Git default |
| **Collapsible files** | Expand/collapse | N/A | N/A | Auto-collapse generated | Collapse | Sections only |
| **Inline commit** | Yes | N/A | N/A | Web UI | Input box | `i` |
| **Vim navigation** | Yes | N/A | N/A | No | No (ext) | Yes |
| **Structural/AST diff** | No | No | Yes (30+ langs) | No | No | No |

## Detailed Findings

### Finding 1: Lazygit's Staging Model is the TUI Gold Standard
**Source:** [Lazygit GitHub](https://github.com/jesseduffield/lazygit), [Improved diff UX issue](https://github.com/jesseduffield/lazygit/issues/2659)
**Key Points:**
- Three-level staging: file (space in file list) > hunk (`a` in diff) > line (`space` on line)
- Visual range with `v`, then `space` to stage range
- Toggle between staged/unstaged view of the same file
- Click-to-stage being improved (issue #3986) -- click in diff to enter staging with clicked line selected
- Line wrapping support added in staging view (PR #4098)
- Philosophy: zero memorization, everything one keypress away

### Finding 2: Delta's Display Engine Sets the Visual Standard
**Source:** [Delta GitHub](https://github.com/dandavison/delta), [Delta docs](https://dandavison.github.io/delta/introduction.html)
**Key Points:**
- Levenshtein edit inference for word-level highlighting (buffers nearby lines for comparison)
- Syntax highlighting using bat's theme engine (200+ themes)
- Side-by-side with line numbers on by default
- Navigate mode (`n`/`N`) for jumping between files in large diffs
- Copy-friendly: `-`/`+` markers removed by default so code can be copied directly
- Hyperlinks for commits linking to GitHub/GitLab

### Finding 3: Difftastic Shows the Future is AST-Aware
**Source:** [Difftastic GitHub](https://github.com/Wilfred/difftastic)
**Key Points:**
- Parses code with tree-sitter, compares ASTs using Dijkstra's algorithm
- Understands reformatting (line splits) vs semantic changes
- Whitespace-aware (knows when whitespace matters, e.g., Python, string literals)
- 30+ language support
- Falls back to line-oriented diff on parse errors
- Not line-oriented: shows what actually changed structurally
- Limitation: output is for humans only, does not generate patches

### Finding 4: GitHub's 2025 Files Changed Redesign
**Source:** [GitHub Changelog](https://github.blog/changelog/2025-06-26-improved-pull-request-files-changed-experience-now-in-public-preview/)
**Key Points:**
- Faster diff rendering with lower memory usage
- Resizable file tree
- Side panel for discovering comments with search
- Comment/annotation indicators in file tree
- Pending comments shown in review submission panel
- Local draft comments persist across page refreshes
- Seamless toggle between split and unified views
- `i` key to toggle comment minimization
- Keyboard navigation and screen reader landmarks for accessibility

### Finding 5: VS Code's Gutter-Based Inline Actions
**Source:** [VS Code Source Control docs](https://code.visualstudio.com/docs/sourcecontrol/overview)
**Key Points:**
- Quick diff gutter decorations show changes inline in the editor
- Click gutter decoration to see inline diff
- Stage/unstage specific lines from diff view gutter
- Side-by-side or inline diff view toggle
- AI-powered commit message generation from staged changes
- F7/Shift+F7 navigation between changes

## Recommendations for gdiff

### High Priority (Strong ROI)

1. **Add side-by-side mode toggle** -- Every mature diff tool supports this. A keybinding (e.g., `d` for "display mode") to toggle between unified and side-by-side would match user expectations from GitHub, VS Code, and delta.

2. **Expandable context lines** -- Allow users to press a key (e.g., `+`/`-` or `=`) to show more/fewer context lines around hunks. GitHub's "expand" pattern is well-understood. Start with 3 (current), allow expanding to full file context.

3. **Syntax highlighting in diff view** -- Delta and difftastic have made syntax-highlighted diffs the expectation. Use a Go syntax highlighting library (e.g., chroma) to colorize code in the diff pane. This is the single biggest visual quality improvement available.

4. **Visual feedback for staging operations** -- Brief flash/highlight when a line/hunk is staged (similar to vim's yank highlight). Makes the staging feel responsive and confirms the action.

### Medium Priority (Polish)

5. **Dual line numbers** -- Show both old and new file line numbers. Currently common in delta, GitHub, and VS Code. Critical for referencing specific lines during review.

6. **File tree change indicators** -- Show `M` (modified), `A` (added), `D` (deleted), `R` (renamed) next to files in the tree, plus counts of additions/deletions (e.g., `+12 -3`). GitHub and lazygit both do this.

7. **Copy-friendly diff output** -- When selecting text, strip the `+`/`-` markers automatically (delta does this). If the TUI supports clipboard, this is a nice touch.

8. **Hunk header with function context** -- Show the function/method name in hunk headers (git already provides this with `--function-context`). Helps orient the reviewer in large files.

### Lower Priority (Differentiation)

9. **Collapsible file sections** -- Auto-collapse generated files, lock files, and binary files. Allow manual collapse per file. GitHub and GitLab both do this.

10. **Diff statistics bar** -- Show a minimap-style bar or sparkline showing where changes are concentrated in the file. Novel for TUI tools.

11. **Side panel summary** -- Show total files changed, lines added/removed, and a progress indicator of reviewed files. GitHub's new 2025 redesign emphasizes this.

12. **Word-level highlighting upgrade** -- gdiff already has LCS character-level detection. Consider adding Levenshtein-based word boundary detection (like delta) as an alternative mode, which tends to produce more readable highlights for prose/comments.

### Implementation Notes

- **Side-by-side in terminal**: Requires calculating available width, splitting the viewport. Ratatui/Lip Gloss can handle this, but consider minimum terminal width requirements (120+ cols recommended for side-by-side).
- **Syntax highlighting**: The Go `chroma` library supports 200+ languages and integrates well with terminal output. It adds a dependency but the visual impact is substantial.
- **Context expansion**: Requires re-running `git diff -U<n>` with increased context count, or loading the full file and merging with diff hunks. The latter is more responsive.
- **Staging feedback**: A simple 100ms highlight on the staged line(s) using a distinct background color, then fade to the "staged" color, provides excellent tactile feedback.

### What gdiff Already Does Well

- **Character-level staging** via visual mode is a genuine differentiator. No other TUI tool does this (lazygit stops at line level).
- **Vim-style navigation** is comprehensive with `gg`, `G`, `Ctrl+d/u`, and hunk jumping.
- **Inline commit** with `i` matches lazygit's workflow.
- **Hunk navigation** with `}`/`{` and `]c`/`[c` covers both vim-diff conventions.
- **Revert with confirmation** (`x`) is a safety feature many tools lack.
- **Async operations with spinners** prevents UI blocking.

### Competitive Positioning

gdiff occupies a unique niche: **lazygit-class staging UX with character-level granularity**. To strengthen this position:
1. Match lazygit's display quality (syntax highlighting, line numbers)
2. Match delta's visual polish (word-level highlighting, themes)
3. Keep the character-level staging as the primary differentiator
4. Add side-by-side mode to match GitHub/VS Code expectations

## Sources

1. [Lazygit GitHub](https://github.com/jesseduffield/lazygit) - TUI git client with staging workflow
2. [Lazygit Improved Diff UX Issue #2659](https://github.com/jesseduffield/lazygit/issues/2659) - Discussion on diff UX improvements
3. [Lazygit Click-to-Stage Issue #3986](https://github.com/jesseduffield/lazygit/issues/3986) - Mouse interaction in diff view
4. [Delta GitHub](https://github.com/dandavison/delta) - Syntax-highlighting pager for git
5. [Delta Documentation](https://dandavison.github.io/delta/introduction.html) - Full feature reference
6. [Difftastic GitHub](https://github.com/Wilfred/difftastic) - AST-aware structural diff tool
7. [GitHub Files Changed Redesign (2025)](https://github.blog/changelog/2025-06-26-improved-pull-request-files-changed-experience-now-in-public-preview/) - New PR review experience
8. [GitLab MR Changes Docs](https://docs.gitlab.com/user/project/merge_requests/changes/) - Diff view features
9. [VS Code Source Control](https://code.visualstudio.com/docs/sourcecontrol/overview) - Inline diff and staging features
10. [VS Code Staging & Commits](https://code.visualstudio.com/docs/sourcecontrol/staging-commits) - Line-level staging
11. [The (lazy) Git UI You Didn't Know You Need](https://www.bwplotka.dev/2025/lazygit/) - Lazygit philosophy and workflow
12. [Viget: Word-Level Git Highlights](https://www.viget.com/articles/dress-up-your-git-diffs-with-word-level-highlights/) - Character/word diff approaches
13. [Git diff-highlight](https://github.com/git/git/tree/master/contrib/diff-highlight) - Built-in sub-word highlighting
14. [GNU Diff Options](https://www.gnu.org/software/diffutils/manual/html_node/diff-Options.html) - Context line standards

## Open Questions

- Should side-by-side mode share the same staging keybindings or have adapted ones for the two-pane layout?
- Is tree-sitter integration (for AST-aware diffs like difftastic) worth the complexity for a staging-focused tool?
- Should gdiff support delta as an optional external pager for rendering, or keep rendering self-contained?
