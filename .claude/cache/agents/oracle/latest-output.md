# Research Report: Charmbracelet Lip Gloss & Bubbles for Diff Tool
Generated: 2026-01-26

## Summary

Lip Gloss provides CSS-like terminal styling with automatic color profile detection, adaptive theming, and layout utilities ideal for diff highlighting. Bubbles offers pre-built TUI components (viewport, list, textinput, help, key, spinner, progress) that directly map to diff tool requirements. Both integrate seamlessly with the Bubble Tea framework using the Elm Architecture pattern.

## Questions Answered

### Q1: Lip Gloss Styling Capabilities for Diff Highlighting
**Answer:** Lip Gloss supports comprehensive text styling including Bold, Italic, Faint, Strikethrough, Underline, and Reverse - all useful for diff highlighting (additions, deletions, changes). Color support is robust with automatic terminal detection and degradation.
**Source:** https://github.com/charmbracelet/lipgloss
**Confidence:** High

### Q2: Lip Gloss Color Schemes and Theming
**Answer:** AdaptiveColor automatically selects colors based on terminal background (light/dark). CompleteAdaptiveColor allows specifying exact values for TrueColor, ANSI256, and ANSI modes per light/dark theme.
**Source:** https://pkg.go.dev/github.com/charmbracelet/lipgloss
**Confidence:** High

### Q3: Lip Gloss Border and Layout Options
**Answer:** Built-in borders: NormalBorder, RoundedBorder, ThickBorder, DoubleBorder, BlockBorder, HiddenBorder. Layout utilities: JoinHorizontal, JoinVertical, Place, PlaceHorizontal, PlaceVertical. Custom borders supported.
**Source:** https://github.com/charmbracelet/lipgloss/blob/master/borders.go
**Confidence:** High

### Q4: Lip Gloss Rendering Performance
**Answer:** No specific benchmarks found for Go implementation. Rust port claims 10-100x faster style comparisons. The library auto-detects terminal capabilities and strips ANSI when not outputting to TTY, which is performance-conscious design.
**Source:** https://github.com/whit3rabbit/lipgloss-rs
**Confidence:** Medium (inferred from design)

### Q5: Bubbles Viewport Component
**Answer:** Supports vertical scrolling with optional pager keybindings and mouse wheel support. High-performance mode available (deprecated) for alternate screen buffer apps. Methods: PageDown, PageUp, HalfPageDown, ScrollDown. MouseWheelDelta configurable.
**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport
**Confidence:** High

### Q6: Bubbles TextInput Component
**Answer:** Full text input with suggestions support (MatchedSuggestions, CurrentSuggestionIndex). Widely adopted (1,797 importers). Good for commit message entry.
**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/textinput
**Confidence:** High

### Q7: Bubbles List Component
**Answer:** Feature-rich browsing component with optional filtering, pagination, help, status messages, and activity spinner. All features can be enabled/disabled independently.
**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/list
**Confidence:** High

### Q8: Bubbles Help Component
**Answer:** Generates help views from keybindings. Implements KeyMap interface with ShortHelp() and FullHelp() methods. Auto-hides disabled keybindings. Width-aware truncation.
**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/help
**Confidence:** High

### Q9: Bubbles Key Component
**Answer:** Non-visual keybinding manager. key.NewBinding with WithKeys and WithHelp. key.Matches(msg, binding) for handling. SetEnabled for dynamic enable/disable. Integrates with help component.
**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/key
**Confidence:** High

### Q10: Bubbles Spinner/Progress Components
**Answer:** Spinner has multiple built-in frames and custom frame support. Progress supports solid/gradient fills with spring animation (via Harmonica). Both useful for operations like staging, diffing large repos.
**Source:** https://github.com/charmbracelet/bubbles
**Confidence:** High

### Q11: Component Customization/Extension
**Answer:** Components are Go structs that can be embedded and extended. All expose Model structs with public fields. View() methods can be overridden or wrapped.
**Source:** https://donderom.com/posts/managing-nested-models-with-bubble-tea/
**Confidence:** High

### Q12: Component Composition Patterns
**Answer:** Parent embeds sub-models as fields. Update() delegates messages to children. View() composes child views. WindowSizeMsg should be routed to all children. Tree structure with root as message router.
**Source:** https://deepwiki.com/charmbracelet/bubbletea/6.5-component-integration
**Confidence:** High

## Detailed Findings

### Finding 1: Diff Styling with Lip Gloss

**Source:** https://github.com/charmbracelet/lipgloss

**Key Points:**
- Text formatting: Bold, Italic, Faint, Strikethrough, Underline, Reverse
- Block formatting: Padding (Top/Right/Bottom/Left), Margins
- Tab handling: Converts tabs to 4 spaces (configurable via TabWidth)
- Color auto-degradation: TrueColor -> ANSI256 -> ANSI -> no color

**Code Example:**
```go
// Diff line styles
var (
    addedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "#22863a", Dark: "#85e89d"}).
        Bold(true)
    
    removedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "#cb2431", Dark: "#f97583"}).
        Strikethrough(true)
    
    contextStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "#6a737d", Dark: "#959da5"})
    
    hunkHeaderStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "#6f42c1", Dark: "#b392f0"}).
        Background(lipgloss.AdaptiveColor{Light: "#f6f8fa", Dark: "#24292e"})
)
```

### Finding 2: Layout for Side-by-Side Diff

**Source:** https://pkg.go.dev/github.com/charmbracelet/lipgloss

**Key Points:**
- JoinHorizontal(position, ...strings) - aligns with Top, Center, Bottom
- JoinVertical(position, ...strings) - aligns with Left, Center, Right
- Place(width, height, hPos, vPos, str) - positions within box
- Width/Height constraints via MaxWidth, MaxHeight

**Code Example:**
```go
// Side-by-side diff layout
func renderSideBySide(left, right string, width int) string {
    halfWidth := width / 2
    
    leftPanel := lipgloss.NewStyle().
        Width(halfWidth).
        Border(lipgloss.NormalBorder()).
        BorderRight(false).
        Render(left)
    
    rightPanel := lipgloss.NewStyle().
        Width(halfWidth).
        Border(lipgloss.NormalBorder()).
        Render(right)
    
    return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}
```

### Finding 3: Viewport for Scrolling Diff Content

**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport

**Key Points:**
- SetContent(string) to load diff content
- GotoTop(), GotoBottom() for navigation
- AtTop(), AtBottom() for boundary detection
- MouseWheelEnabled, MouseWheelDelta for mouse support
- YOffset for current scroll position

**Code Example:**
```go
func newDiffViewport(width, height int) viewport.Model {
    vp := viewport.New(width, height)
    vp.MouseWheelEnabled = true
    vp.MouseWheelDelta = 3
    return vp
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "g":
            m.viewport.GotoTop()
        case "G":
            m.viewport.GotoBottom()
        }
    }
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}
```

### Finding 4: File List for Selection

**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/list

**Key Points:**
- Filtering built-in (fuzzy search)
- Pagination for large file lists
- Custom item rendering via ItemDelegate
- Status messages for feedback
- Spinner for loading states

**Code Example:**
```go
type fileItem struct {
    path     string
    status   string // "M", "A", "D", "R"
    additions int
    deletions int
}

func (i fileItem) Title() string       { return i.path }
func (i fileItem) Description() string { return fmt.Sprintf("+%d -%d", i.additions, i.deletions) }
func (i fileItem) FilterValue() string { return i.path }

func newFileList(files []fileItem, width, height int) list.Model {
    l := list.New(files, list.NewDefaultDelegate(), width, height)
    l.Title = "Changed Files"
    l.SetFilteringEnabled(true)
    return l
}
```

### Finding 5: Help and Keybindings

**Source:** https://pkg.go.dev/github.com/charmbracelet/bubbles/help

**Key Points:**
- KeyMap interface: ShortHelp() []key.Binding, FullHelp() [][]key.Binding
- Width-aware rendering (auto-truncation)
- ShowAll toggle for full/short view
- Disabled bindings auto-hidden

**Code Example:**
```go
type keyMap struct {
    Up       key.Binding
    Down     key.Binding
    Toggle   key.Binding
    Stage    key.Binding
    Unstage  key.Binding
    Quit     key.Binding
    Help     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down},
        {k.Stage, k.Unstage, k.Toggle},
        {k.Help, k.Quit},
    }
}

var keys = keyMap{
    Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
    Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
    Toggle:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "toggle view")),
    Stage:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "stage")),
    Unstage: key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unstage")),
    Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
    Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}
```

### Finding 6: Component Composition Pattern

**Source:** https://deepwiki.com/charmbracelet/bubbletea/6.5-component-integration

**Key Points:**
- Parent model embeds child models as fields
- Init() calls child Init() methods
- Update() delegates to relevant children
- View() composes child View() outputs
- WindowSizeMsg must propagate to all children

**Code Example:**
```go
type Model struct {
    // Sub-components
    fileList list.Model
    viewport viewport.Model
    textInput textinput.Model
    help     help.Model
    
    // App state
    keys     keyMap
    width    int
    height   int
    mode     viewMode
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Propagate to all children
        m.fileList.SetSize(msg.Width/3, msg.Height-2)
        m.viewport.Width = msg.Width*2/3
        m.viewport.Height = msg.Height - 2
    
    case tea.KeyMsg:
        if key.Matches(msg, m.keys.Quit) {
            return m, tea.Quit
        }
    }
    
    // Delegate to active component
    var cmd tea.Cmd
    switch m.mode {
    case fileListMode:
        m.fileList, cmd = m.fileList.Update(msg)
    case diffMode:
        m.viewport, cmd = m.viewport.Update(msg)
    case commitMode:
        m.textInput, cmd = m.textInput.Update(msg)
    }
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m Model) View() string {
    leftPanel := m.fileList.View()
    rightPanel := m.viewport.View()
    
    main := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
    help := m.help.View(m.keys)
    
    return lipgloss.JoinVertical(lipgloss.Left, main, help)
}
```

## Comparison Matrix

| Component | Diff Tool Use Case | Customization | Complexity |
|-----------|-------------------|---------------|------------|
| Viewport | Scrolling diff content | High (any content) | Low |
| List | File selection | High (ItemDelegate) | Medium |
| TextInput | Commit messages | Medium (suggestions) | Low |
| Help | Keybinding display | High (KeyMap interface) | Low |
| Key | Input handling | High (custom bindings) | Low |
| Spinner | Loading operations | Low (frames only) | Low |
| Progress | Staging progress | High (colors, animation) | Low |

| Border Style | Visual | Use Case |
|--------------|--------|----------|
| NormalBorder | Simple lines | General panels |
| RoundedBorder | Rounded corners | Modern aesthetic |
| ThickBorder | Heavy lines | Emphasis |
| DoubleBorder | Double lines | Classic look |
| HiddenBorder | Invisible | Spacing without lines |

## Recommendations

### For This Diff Tool

1. **Use AdaptiveColor for all diff styling** - Ensures readability in both light and dark terminals without user configuration.

2. **Viewport for diff display** - Standard pager keybindings (j/k, Page Up/Down) are pre-built. Mouse wheel support is optional flag.

3. **List for file selection** - Built-in filtering is valuable for large diffs. Custom ItemDelegate can show +/- counts and status icons.

4. **Composable KeyMap** - Define all keybindings in one struct, integrate with help component for auto-generated help text.

5. **JoinHorizontal for side-by-side** - Natural fit for side-by-side diff view. Use Border on panels for separation.

### Implementation Notes

- **WindowSizeMsg handling is critical** - Must propagate to all sub-components for proper resizing
- **High-performance rendering is deprecated** - Use standard Bubble Tea renderer
- **Tab character handling** - Lip Gloss converts tabs to 4 spaces; consider this for indentation in diffs
- **Color degradation is automatic** - No need to handle different terminal capabilities manually
- **Disabled keybindings auto-hide in help** - Use SetEnabled(false) for context-sensitive help

## Example: Diff Tool Architecture

```
┌─────────────────────────────────────────────────────────┐
│ Model                                                   │
│ ├── list.Model (file list)                              │
│ ├── viewport.Model (diff view)                          │
│ ├── textinput.Model (commit message)                    │
│ ├── help.Model (keybinding help)                        │
│ ├── keyMap (all keybindings)                            │
│ └── state (mode, selected file, etc.)                   │
├─────────────────────────────────────────────────────────┤
│ Lip Gloss Styles                                        │
│ ├── addedStyle (green, bold)                            │
│ ├── removedStyle (red, strikethrough)                   │
│ ├── contextStyle (gray)                                 │
│ ├── hunkHeaderStyle (purple background)                 │
│ ├── panelStyle (borders, padding)                       │
│ └── statusBarStyle (inverted colors)                    │
├─────────────────────────────────────────────────────────┤
│ Layout (View function)                                  │
│ ┌───────────────┬───────────────────────────────┐       │
│ │ File List     │ Diff Viewport                 │       │
│ │ (list.Model)  │ (viewport.Model)              │       │
│ │               │                               │       │
│ │ M file1.go    │ @@ -10,5 +10,7 @@             │       │
│ │ A file2.go    │  context line                 │       │
│ │ D file3.go    │ -removed line                 │       │
│ │               │ +added line                   │       │
│ ├───────────────┴───────────────────────────────┤       │
│ │ Help: ↑/k up  ↓/j down  s stage  ? help  q quit│       │
│ └────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────┘
```

## Sources

1. [Lip Gloss GitHub](https://github.com/charmbracelet/lipgloss) - Main repository with examples
2. [Lip Gloss Go Docs](https://pkg.go.dev/github.com/charmbracelet/lipgloss) - API reference
3. [Bubbles GitHub](https://github.com/charmbracelet/bubbles) - TUI component library
4. [Viewport Package](https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport) - Scrolling component
5. [List Package](https://pkg.go.dev/github.com/charmbracelet/bubbles/list) - File selection component
6. [TextInput Package](https://pkg.go.dev/github.com/charmbracelet/bubbles/textinput) - Text entry component
7. [Help Package](https://pkg.go.dev/github.com/charmbracelet/bubbles/help) - Help display component
8. [Key Package](https://pkg.go.dev/github.com/charmbracelet/bubbles/key) - Keybinding management
9. [Managing Nested Models](https://donderom.com/posts/managing-nested-models-with-bubble-tea/) - Composition patterns
10. [Component Integration](https://deepwiki.com/charmbracelet/bubbletea/6.5-component-integration) - Integration patterns
11. [diffnav](https://github.com/dlvhdr/diffnav) - Reference implementation of diff viewer with Bubble Tea

## Open Questions

- Exact performance benchmarks for Lip Gloss rendering with large diffs (1000+ lines)
- Best practices for lazy loading diff content in viewport
- Optimal approach for synchronized scrolling in side-by-side view
