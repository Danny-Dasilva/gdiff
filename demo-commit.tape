# VHS Demo for gdiff - Commit Workflows
# Demonstrates inline commit and full commit dialog
# Run with: PATH="/tmp:$PATH" vhs demo-commit.tape

Output demo-commit.gif

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

# === Stage Some Files First ===
# Stage a couple files for the commit demo
Type "j"
Sleep 300ms
Type "a"
Sleep 500ms
Type "j"
Sleep 300ms
Type "a"
Sleep 500ms
Type "j"
Sleep 300ms
Type "a"
Sleep 1s

# === Method 1: Inline Commit with 'i' ===
# The inline commit input is at the bottom of the screen
# Press 'i' to focus it
Type "i"
Sleep 800ms

# Type your commit message directly
Type "feat: add logger and update configuration"
Sleep 1s

# Press Enter to commit
Enter
Sleep 2s

# Commit is created! The staged files are now committed
Sleep 1s

# === Setup for Second Demo ===
# Stage more files for the full commit dialog demo
Type "j"
Sleep 300ms
Type "a"
Sleep 500ms
Type "j"
Sleep 300ms
Type "a"
Sleep 1s

# === Method 2: Full Commit Dialog with 'c' ===
# Press 'c' to open the full commit dialog
Type "c"
Sleep 1s

# The dialog shows staged files and allows multi-line messages
# Type a commit message
Type "refactor: improve utility functions"
Sleep 500ms

# You can add more lines in the dialog
Enter
Sleep 200ms
Type "- Add TrimSpace to string functions"
Sleep 500ms
Enter
Sleep 200ms
Type "- Update README with new features"
Sleep 800ms

# Press Enter to commit (or Escape to cancel)
Enter
Sleep 2s

# === Verify Commits ===
# Both commits are now in git history
# The file tree shows no more staged files
Sleep 1.5s

# === Show Help for Commit Keys ===
Type "?"
Sleep 2s
Escape
Sleep 500ms

# Quit
Type "q"
Sleep 500ms
