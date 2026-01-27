package git

import (
	"context"

	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

// GetFileDiff returns the diff for a specific file
func GetFileDiff(ctx context.Context, path string, staged bool) ([]diff.FileDiff, error) {
	args := []string{"diff", "--histogram", "--no-color"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)

	out, err := RunGitCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	return diff.Parse(out), nil
}

// GetAllDiffs returns diffs for all changed files
func GetAllDiffs(ctx context.Context, staged bool) ([]diff.FileDiff, error) {
	args := []string{"diff", "--histogram", "--no-color"}
	if staged {
		args = append(args, "--cached")
	}

	out, err := RunGitCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	return diff.Parse(out), nil
}

// GetDiffStats returns quick stats for all changed files
func GetDiffStats(ctx context.Context, staged bool) (map[string][2]int, error) {
	args := []string{"diff", "--numstat"}
	if staged {
		args = append(args, "--cached")
	}

	out, err := RunGitCommand(ctx, args...)
	if err != nil {
		return nil, err
	}

	return parseNumstat(out), nil
}

func parseNumstat(output string) map[string][2]int {
	stats := make(map[string][2]int)
	// Format: added\tremoved\tpath
	// Binary files show as -\t-\tpath
	lines := splitLines(output)

	for _, line := range lines {
		if line == "" {
			continue
		}

		var added, removed int
		var path string

		// Try to parse numeric values
		n, _ := scanNumstat(line, &added, &removed, &path)
		if n >= 3 && path != "" {
			stats[path] = [2]int{added, removed}
		}
	}

	return stats
}

func scanNumstat(line string, added, removed *int, path *string) (int, error) {
	// Manual parsing since fmt.Sscanf doesn't handle tabs well
	fields := splitTabs(line)
	if len(fields) < 3 {
		return 0, nil
	}

	// Parse added (may be "-" for binary)
	if fields[0] != "-" {
		var a int
		for _, c := range fields[0] {
			if c >= '0' && c <= '9' {
				a = a*10 + int(c-'0')
			}
		}
		*added = a
	}

	// Parse removed (may be "-" for binary)
	if fields[1] != "-" {
		var r int
		for _, c := range fields[1] {
			if c >= '0' && c <= '9' {
				r = r*10 + int(c-'0')
			}
		}
		*removed = r
	}

	*path = fields[2]
	return 3, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitTabs(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}
