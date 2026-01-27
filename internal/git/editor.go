package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetEditor returns the user's preferred editor from environment variables.
// It checks VISUAL first, then EDITOR, falling back to "vi" if neither is set.
func GetEditor() string {
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

// CreateTempCommitFile creates a temporary file with the given initial content
// for editing a commit message. Returns the path to the temp file.
func CreateTempCommitFile(initial string) (string, error) {
	tmpfile, err := os.CreateTemp("", "COMMIT_EDITMSG*.txt")
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.WriteString(initial); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		return "", err
	}

	return tmpfile.Name(), nil
}

// ReadTempCommitFile reads the content from the temp file and returns it.
func ReadTempCommitFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// OpenEditor opens the specified file in the user's preferred editor.
// This function blocks until the editor closes.
// Returns the updated content from the file after the editor closes.
func OpenEditor(filePath string) (string, error) {
	editor := GetEditor()

	// Handle editor commands that might have arguments (e.g., "code --wait")
	parts := strings.Fields(editor)
	cmdName := parts[0]
	args := append(parts[1:], filePath)

	cmd := exec.Command(cmdName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return ReadTempCommitFile(filePath)
}

// EditCommitMessage opens an editor for the user to edit a commit message.
// It creates a temp file with the initial content, opens the editor,
// and returns the edited content. The temp file is cleaned up automatically.
func EditCommitMessage(initial string) (string, error) {
	tmpPath, err := CreateTempCommitFile(initial)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpPath)

	content, err := OpenEditor(tmpPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(content), nil
}

// EditorCmd creates an exec.Cmd for opening the editor with the given file.
// This is useful for bubbletea's ExecProcess command.
func EditorCmd(filePath string) *exec.Cmd {
	editor := GetEditor()

	// Handle editor commands that might have arguments (e.g., "code --wait")
	parts := strings.Fields(editor)
	cmdName := parts[0]
	args := append(parts[1:], filePath)

	// Resolve the editor path
	cmdPath, err := exec.LookPath(cmdName)
	if err != nil {
		// Fall back to using the name directly
		cmdPath = cmdName
	}

	cmd := exec.Command(cmdPath, args...)
	return cmd
}

// GetEditorTempPath returns the path to the git directory's COMMIT_EDITMSG file
// or creates a temp file if not in a git repo.
func GetEditorTempPath() (string, bool) {
	// Try to find .git directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}

	gitDir := filepath.Join(cwd, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		return filepath.Join(gitDir, "COMMIT_EDITMSG"), true
	}

	return "", false
}
