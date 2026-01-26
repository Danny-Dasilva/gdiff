package git

import (
	"context"
	"strings"

	"github.com/danny/gdiff/pkg/diff"
)

// GetStatus returns the list of changed files using porcelain v2 format
func GetStatus(ctx context.Context) ([]diff.FileEntry, error) {
	out, err := RunGitCommand(ctx, "status", "--porcelain=v2", "-z")
	if err != nil {
		return nil, err
	}

	return parseStatusV2(out), nil
}

// parseStatusV2 parses git status --porcelain=v2 -z output
func parseStatusV2(output string) []diff.FileEntry {
	var entries []diff.FileEntry

	// Split by null byte
	parts := strings.Split(output, "\x00")

	for i := 0; i < len(parts); i++ {
		line := parts[i]
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "1 "):
			// Ordinary changed entry: 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
			entry := parseOrdinaryEntry(line)
			if entry != nil {
				entries = append(entries, *entry)
			}

		case strings.HasPrefix(line, "2 "):
			// Renamed/copied entry: 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
			// The original path follows in the next null-separated field
			if i+1 < len(parts) {
				entry := parseRenamedEntry(line, parts[i+1])
				if entry != nil {
					entries = append(entries, *entry)
				}
				i++ // Skip the original path
			}

		case strings.HasPrefix(line, "? "):
			// Untracked: ? <path>
			path := strings.TrimPrefix(line, "? ")
			entries = append(entries, diff.FileEntry{
				Path:        path,
				Status:      diff.StatusUntracked,
				IndexStatus: diff.StatusUnmodified,
				WorkStatus:  diff.StatusUntracked,
			})

		case strings.HasPrefix(line, "! "):
			// Ignored: ! <path>
			path := strings.TrimPrefix(line, "! ")
			entries = append(entries, diff.FileEntry{
				Path:        path,
				Status:      diff.StatusIgnored,
				IndexStatus: diff.StatusUnmodified,
				WorkStatus:  diff.StatusIgnored,
			})

		case strings.HasPrefix(line, "u "):
			// Unmerged: u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
			entry := parseUnmergedEntry(line)
			if entry != nil {
				entries = append(entries, *entry)
			}
		}
	}

	return entries
}

func parseOrdinaryEntry(line string) *diff.FileEntry {
	// Format: 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
	fields := strings.SplitN(line, " ", 9)
	if len(fields) < 9 {
		return nil
	}

	xy := fields[1]
	path := fields[8]

	indexStatus := charToStatus(xy[0])
	workStatus := charToStatus(xy[1])

	// Determine overall status
	status := workStatus
	if indexStatus != diff.StatusUnmodified {
		status = indexStatus
	}

	return &diff.FileEntry{
		Path:        path,
		Status:      status,
		Staged:      indexStatus != diff.StatusUnmodified,
		IndexStatus: indexStatus,
		WorkStatus:  workStatus,
	}
}

func parseRenamedEntry(line, origPath string) *diff.FileEntry {
	// Format: 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path>
	fields := strings.SplitN(line, " ", 10)
	if len(fields) < 10 {
		return nil
	}

	xy := fields[1]
	path := fields[9]

	indexStatus := charToStatus(xy[0])
	workStatus := charToStatus(xy[1])

	return &diff.FileEntry{
		Path:        path,
		OldPath:     origPath,
		Status:      diff.StatusRenamed,
		Staged:      indexStatus != diff.StatusUnmodified,
		IndexStatus: indexStatus,
		WorkStatus:  workStatus,
	}
}

func parseUnmergedEntry(line string) *diff.FileEntry {
	// Format: u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
	fields := strings.SplitN(line, " ", 11)
	if len(fields) < 11 {
		return nil
	}

	path := fields[10]

	return &diff.FileEntry{
		Path:        path,
		Status:      diff.StatusUnmerged,
		IndexStatus: diff.StatusUnmerged,
		WorkStatus:  diff.StatusUnmerged,
	}
}

func charToStatus(c byte) diff.FileStatus {
	switch c {
	case 'M':
		return diff.StatusModified
	case 'A':
		return diff.StatusAdded
	case 'D':
		return diff.StatusDeleted
	case 'R':
		return diff.StatusRenamed
	case 'C':
		return diff.StatusCopied
	case 'U':
		return diff.StatusUnmerged
	case '?':
		return diff.StatusUntracked
	case '!':
		return diff.StatusIgnored
	default:
		return diff.StatusUnmodified
	}
}
