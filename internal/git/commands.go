package git

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// RunGitCommand executes a git command and returns stdout
func RunGitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Include stderr in error for debugging
		if stderr.Len() > 0 {
			return "", &GitError{
				Command: strings.Join(args, " "),
				Stderr:  stderr.String(),
				Err:     err,
			}
		}
		return "", err
	}

	return stdout.String(), nil
}

// GitError wraps git command errors with context
type GitError struct {
	Command string
	Stderr  string
	Err     error
}

func (e *GitError) Error() string {
	if e.Stderr != "" {
		return "git " + e.Command + ": " + e.Stderr
	}
	return "git " + e.Command + ": " + e.Err.Error()
}

func (e *GitError) Unwrap() error {
	return e.Err
}

// IsGitRepo checks if current directory is a git repository
func IsGitRepo(ctx context.Context) bool {
	_, err := RunGitCommand(ctx, "rev-parse", "--git-dir")
	return err == nil
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot(ctx context.Context) (string, error) {
	out, err := RunGitCommand(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(ctx context.Context) (string, error) {
	out, err := RunGitCommand(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
