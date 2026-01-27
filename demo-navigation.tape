# VHS Demo for gdiff - Navigation Features
# Demonstrates all navigation keybindings
# Run with: PATH="/tmp:$PATH" vhs demo-navigation.tape

Output demo-navigation.gif

Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set PlaybackSpeed 0.5

# Setup: run setup-demo.sh first

Type "cd /tmp/gdiff-demo && /tmp/gdiff"
Enter
Sleep 2s

# === File Tree Navigation ===
# j/k - Move down/up through files
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "k"
Sleep 300ms
Type "k"
Sleep 300ms

# gg - Jump to top
Type "gg"
Sleep 600ms

# G - Jump to bottom
Type "G"
Sleep 600ms

# gg back to top
Type "gg"
Sleep 600ms

# === Switch to Diff Pane ===
# Tab - Switch between panes
Tab
Sleep 800ms

# === Diff View Navigation ===
# j/k - Line by line
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "k"
Sleep 200ms
Type "k"
Sleep 200ms

# Ctrl+d - Half page down
Ctrl+d
Sleep 600ms
Ctrl+d
Sleep 600ms

# Ctrl+u - Half page up
Ctrl+u
Sleep 600ms
Ctrl+u
Sleep 600ms

# } - Jump to next hunk
Type "}"
Sleep 500ms
Type "}"
Sleep 500ms
Type "}"
Sleep 500ms

# { - Jump to previous hunk
Type "{"
Sleep 500ms
Type "{"
Sleep 500ms

# gg/G in diff view
Type "gg"
Sleep 600ms
Type "G"
Sleep 600ms
Type "gg"
Sleep 600ms

# === Switch Back to File Tree ===
Tab
Sleep 600ms

# Navigate to different file to show diff updates
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Switch to diff to see new file's diff
Tab
Sleep 1s

# Scroll through this file's diff
Ctrl+d
Sleep 500ms
Type "}"
Sleep 500ms

# === Final ===
Sleep 1s
Type "q"
Sleep 500ms
