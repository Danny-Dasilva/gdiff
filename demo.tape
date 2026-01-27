# VHS Demo for gdiff - Git Diff TUI
# Full workflow: view diff → navigate → stage → commit
# Run with: vhs demo.tape

Output demo.gif

Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set PlaybackSpeed 0.5

# Start gdiff
Type "./gdiff"
Enter
Sleep 2s

# === SECTION 1: VS Code-style Interface ===
# Show the new layout with commit input and staged/unstaged sections
Sleep 1.5s

# === SECTION 2: File Navigation ===
# Navigate through the file sections
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Jump to bottom
Type "G"
Sleep 800ms

# Jump back to top
Type "gg"
Sleep 800ms

# === SECTION 3: Diff Viewing ===
# Switch to diff pane
Type "	"
Sleep 1s

# Scroll through the diff
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms

# Half page down
Ctrl+d
Sleep 800ms

# Half page up
Ctrl+u
Sleep 800ms

# Jump between hunks
Type "}"
Sleep 700ms
Type "}"
Sleep 700ms
Type "{"
Sleep 700ms

# === SECTION 4: Staging Files ===
# Go back to file tree
Type "	"
Sleep 500ms

# Navigate to a file and stage it
Type "j"
Sleep 300ms
Type "a"
Sleep 1s

# Stage another file
Type "j"
Sleep 300ms
Type "a"
Sleep 1s

# === SECTION 5: Hunk Staging ===
# Go to diff view
Type "	"
Sleep 500ms

# Stage a hunk
Type "S"
Sleep 1s

# === SECTION 6: Visual Mode ===
# Enter visual mode for character selection
Type "v"
Sleep 600ms

# Select characters
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms

# Stage selection
Type "s"
Sleep 1s

# === SECTION 7: Review Staged ===
# Toggle to staged view
Type "t"
Sleep 1.5s

# Toggle back
Type "t"
Sleep 1s

# === SECTION 8: Inline Commit ===
# Press 'i' to focus commit input
Type "i"
Sleep 800ms

# Type commit message
Type "feat: add new TUI improvements"
Sleep 800ms

# Submit with Enter
Enter
Sleep 1.5s

# === SECTION 9: Help ===
Type "?"
Sleep 2s
Escape
Sleep 500ms

# Quit
Type "q"
Sleep 500ms
