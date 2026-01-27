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

# === SECTION 1: File Navigation ===
# Show the two-pane interface
Sleep 1s

# Navigate through file tree
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Go back up
Type "k"
Sleep 300ms
Type "k"
Sleep 300ms

# Jump to bottom of file list
Type "G"
Sleep 800ms

# Jump back to top
Type "gg"
Sleep 800ms

# === SECTION 2: Diff Viewing ===
# Switch to diff pane to view changes
Type "	"
Sleep 1s

# Scroll through the diff content
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms

# Half page down
Ctrl+d
Sleep 800ms

# Half page down again
Ctrl+d
Sleep 800ms

# Half page up
Ctrl+u
Sleep 800ms

# Jump to next hunk
Type "}"
Sleep 700ms

# Jump to next hunk
Type "}"
Sleep 700ms

# Jump to previous hunk
Type "{"
Sleep 700ms

# === SECTION 3: Staging Files ===
# Go back to file tree
Type "	"
Sleep 500ms

# Navigate to a file
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

# Stage entire file with 'a'
Type "a"
Sleep 1s

# Move to next file
Type "j"
Sleep 300ms

# Stage another file
Type "a"
Sleep 1s

# === SECTION 4: Hunk Staging ===
# Go to diff view
Type "	"
Sleep 500ms

# Navigate to a hunk
Type "}"
Sleep 500ms

# Stage just this hunk with 'S'
Type "S"
Sleep 1s

# === SECTION 5: Visual Mode (Character Selection) ===
# Navigate to a specific line
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms

# Enter visual mode
Type "v"
Sleep 600ms

# Select characters with h/l
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms

# Stage the selection
Type "s"
Sleep 1s

# === SECTION 6: Review Staged Changes ===
# Toggle to staged view
Type "t"
Sleep 1.5s

# Scroll through staged changes
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms

# Toggle back to unstaged
Type "t"
Sleep 1s

# === SECTION 7: Commit Workflow ===
# Open commit dialog
Type "c"
Sleep 1s

# Type commit message
Type "feat: add new TUI components"
Sleep 800ms

# Submit commit with Enter
Enter
Sleep 1.5s

# === SECTION 8: Show Help ===
# Display help overlay
Type "?"
Sleep 2.5s

# Close help
Escape
Sleep 500ms

# === END ===
# Quit the application
Type "q"
Sleep 500ms
