# VHS Demo for gdiff - All Staging Methods
# Demonstrates file, hunk, line, and character staging
# Run with: PATH="/tmp:$PATH" vhs demo-staging.tape

Output demo-staging.gif

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

# === Method 1: File Staging with 'a' ===
# Navigate to a file in the tree
Type "j"
Sleep 400ms

# Stage entire file with 'a'
Type "a"
Sleep 1s

# The file moves to the Staged section
Sleep 500ms

# === Method 2: Hunk Staging with 'S' ===
# Navigate to another file
Type "j"
Sleep 400ms

# Switch to diff view
Tab
Sleep 800ms

# Navigate to a hunk
Type "}"
Sleep 500ms

# Stage entire hunk with 'S'
Type "S"
Sleep 1s

# === Method 3: Line Staging with Visual Mode ===
# Go to a different hunk
Type "}"
Sleep 500ms

# Enter line visual mode with 'V'
Type "V"
Sleep 600ms

# Select multiple lines with j
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

# Stage selected lines with 's'
Type "s"
Sleep 1s

# === Method 4: Character Staging ===
# Navigate to a line with inline changes
Type "gg"
Sleep 500ms
Type "}"
Sleep 500ms

# Enter character visual mode with 'v'
Type "v"
Sleep 600ms

# Use h/l to select specific characters
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

# Stage just those characters with 's'
Type "s"
Sleep 1s

# === Review Staged vs Unstaged ===
# Go back to file tree
Tab
Sleep 500ms

# See the staged section has our changes
Sleep 1s

# Toggle to staged view with 't'
Type "t"
Sleep 1.5s

# See what will be committed
Tab
Sleep 500ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

# Toggle back
Type "t"
Sleep 1s

# === Unstaging ===
# Navigate to staged section
Tab
Sleep 500ms
Type "gg"
Sleep 500ms

# Select a staged file
Type "j"
Sleep 300ms

# Unstage with 'A'
Type "A"
Sleep 1s

# File moves back to unstaged
Sleep 500ms

# Quit
Type "q"
Sleep 500ms
