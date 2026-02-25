# VHS Demo for gdiff - Git Diff TUI
# Full workflow: view diff -> navigate -> stage -> commit
# Run with: PATH="/tmp:$PATH" vhs demo.tape

Output demo.gif

Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set PlaybackSpeed 0.5

# Setup: run setup-demo.sh first to create /tmp/gdiff-demo

# Start gdiff in the demo directory
Type "/tmp/gdiff -C /tmp/gdiff-demo"
Enter
Sleep 2s

# === VS Code-style Interface ===
# The interface shows:
# - Left pane: File tree with Staged/Unstaged sections
# - Right pane: Diff view with syntax highlighting
# - Bottom: Inline commit input
Sleep 1.5s

# === File Navigation ===
# Navigate through files with j/k
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Jump to bottom with G
Type "G"
Sleep 800ms

# Jump back to top with gg
Type "gg"
Sleep 800ms

# === Diff Viewing ===
# Switch to diff pane with Tab
Tab
Sleep 1s

# Scroll through the diff
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms

# Half page down with Ctrl+d
Ctrl+d
Sleep 800ms

# Half page up with Ctrl+u
Ctrl+u
Sleep 800ms

# Jump between hunks with } and {
Type "}"
Sleep 700ms
Type "}"
Sleep 700ms
Type "{"
Sleep 700ms

# === Staging Files ===
# Go back to file tree
Tab
Sleep 500ms

# Navigate to a file and stage it with 'a'
Type "j"
Sleep 300ms
Type "a"
Sleep 1s

# Stage another file
Type "j"
Sleep 300ms
Type "a"
Sleep 1s

# === Hunk Staging ===
# Go to diff view to stage a specific hunk
Tab
Sleep 500ms

# Stage current hunk with S
Type "S"
Sleep 1s

# === Visual Selection ===
# Enter visual mode for line selection
Type "v"
Sleep 600ms

# Select lines with j/k
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

# Stage selection with s
Type "s"
Sleep 1s

# === Review Staged Changes ===
# Toggle to staged view with t
Type "t"
Sleep 1.5s

# Toggle back to unstaged view
Type "t"
Sleep 1s

# === Inline Commit ===
# Press 'i' to focus the commit input
Type "i"
Sleep 800ms

# Type commit message
Type "feat: add logging and improve utilities"
Sleep 800ms

# Submit with Enter (creates the commit)
Enter
Sleep 1.5s

# === Show Help ===
Type "?"
Sleep 2s
Escape
Sleep 500ms

# Quit
Type "q"
Sleep 500ms
