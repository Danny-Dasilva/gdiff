# VHS Demo for gdiff - Character-Level Diff and Staging
# Demonstrates granular character selection and staging
# Run with: PATH="/tmp:$PATH" vhs demo-character-diff.tape

Output demo-character-diff.gif

Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set PlaybackSpeed 0.5

# Setup: run setup-demo.sh first

Type "/tmp/gdiff -C /tmp/gdiff-demo"
Enter
Sleep 2s

# === Navigate to File with Inline Changes ===
# utils.go has small inline changes (added TrimSpace calls)
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

# Switch to diff view
Tab
Sleep 1s

# === View Character-Level Highlighting ===
# The diff shows character-level changes highlighted
# Red/green highlighting shows exactly what changed within lines
Sleep 1.5s

# Navigate to see the inline changes
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms

# === Enter Character Visual Mode ===
# Press 'v' to enter character selection mode
Type "v"
Sleep 800ms

# === Select Specific Characters ===
# Use 'l' to expand selection right (character by character)
Type "l"
Sleep 250ms
Type "l"
Sleep 250ms
Type "l"
Sleep 250ms
Type "l"
Sleep 250ms
Type "l"
Sleep 250ms
Type "l"
Sleep 250ms
Type "l"
Sleep 250ms

# Use 'h' to shrink selection left
Type "h"
Sleep 250ms
Type "h"
Sleep 250ms

# === Stage Selected Characters ===
# Press 's' to stage only the selected characters
Type "s"
Sleep 1.5s

# === Verify Partial Staging ===
# The line is now partially staged
# Toggle to staged view to see what was staged
Type "t"
Sleep 1.5s

# Navigate to see the staged portion
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

Sleep 1s

# Toggle back to unstaged to see remaining changes
Type "t"
Sleep 1.5s

# The unstaged view shows only the remaining characters
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

Sleep 1s

# === Another Character Selection ===
# Navigate to another changed line
Type "}"
Sleep 500ms

# Enter visual mode again
Type "v"
Sleep 600ms

# Select a different range
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms
Type "l"
Sleep 200ms

# Stage it
Type "s"
Sleep 1s

# === Final Review ===
Type "t"
Sleep 1.5s

# Quit
Type "q"
Sleep 500ms
