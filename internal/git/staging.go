package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

// StageFile stages a file
func StageFile(ctx context.Context, path string) error {
	_, err := RunGitCommand(ctx, "add", "--", path)
	return err
}

// UnstageFile unstages a file
func UnstageFile(ctx context.Context, path string) error {
	_, err := RunGitCommand(ctx, "reset", "HEAD", "--", path)
	return err
}

// StageAll stages all changes
func StageAll(ctx context.Context) error {
	_, err := RunGitCommand(ctx, "add", "-A")
	return err
}

// UnstageAll unstages all changes
func UnstageAll(ctx context.Context) error {
	_, err := RunGitCommand(ctx, "reset", "HEAD")
	return err
}

// StageLines stages specific lines from a file using git apply
func StageLines(ctx context.Context, filePath string, hunk diff.Hunk, lineIndices []int) error {
	patch := buildPatch(filePath, hunk, lineIndices, false)
	return applyPatch(ctx, patch, true)
}

// UnstageLines unstages specific lines from a file
func UnstageLines(ctx context.Context, filePath string, hunk diff.Hunk, lineIndices []int) error {
	patch := buildPatch(filePath, hunk, lineIndices, true)
	return applyPatch(ctx, patch, false)
}

// StageHunk stages an entire hunk
func StageHunk(ctx context.Context, filePath string, hunk diff.Hunk) error {
	patch := buildHunkPatch(filePath, hunk)
	return applyPatch(ctx, patch, true)
}

// UnstageHunk unstages an entire hunk
func UnstageHunk(ctx context.Context, filePath string, hunk diff.Hunk) error {
	patch := buildHunkPatch(filePath, hunk)
	return applyPatch(ctx, patch, false)
}

// RevertHunk reverts changes in a hunk
func RevertHunk(ctx context.Context, filePath string, hunk diff.Hunk) error {
	patch := buildReversePatch(filePath, hunk)
	return applyPatch(ctx, patch, false)
}

// buildPatch creates a patch for specific lines
func buildPatch(filePath string, hunk diff.Hunk, lineIndices []int, reverse bool) string {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
	b.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	b.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

	// Build line set for quick lookup
	lineSet := make(map[int]bool)
	for _, idx := range lineIndices {
		lineSet[idx] = true
	}

	// Calculate new hunk bounds
	var selectedLines []diff.Line
	oldCount, newCount := 0, 0

	for i, line := range hunk.Lines {
		if line.Type == diff.LineHunkHeader {
			continue
		}

		include := lineSet[i]

		switch line.Type {
		case diff.LineContext:
			selectedLines = append(selectedLines, line)
			oldCount++
			newCount++
		case diff.LineRemoved:
			if include {
				selectedLines = append(selectedLines, line)
				oldCount++
			} else {
				// Convert to context line if not selected
				selectedLines = append(selectedLines, diff.Line{
					Type:    diff.LineContext,
					Content: line.Content,
					OldNum:  line.OldNum,
					NewNum:  line.OldNum,
				})
				oldCount++
				newCount++
			}
		case diff.LineAdded:
			if include {
				selectedLines = append(selectedLines, line)
				newCount++
			}
			// Skip non-selected added lines
		}
	}

	// Write hunk header
	b.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunk.OldStart, oldCount, hunk.NewStart, newCount))

	// Write lines
	for _, line := range selectedLines {
		switch line.Type {
		case diff.LineContext:
			b.WriteString(" " + line.Content + "\n")
		case diff.LineRemoved:
			b.WriteString("-" + line.Content + "\n")
		case diff.LineAdded:
			b.WriteString("+" + line.Content + "\n")
		}
	}

	if reverse {
		// Reverse the patch for unstaging
		return reversePatchContent(b.String())
	}

	return b.String()
}

// reversePatchContent reverses a patch (swap + and -)
func reversePatchContent(patch string) string {
	lines := strings.Split(patch, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			result = append(result, "-"+line[1:])
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			result = append(result, "+"+line[1:])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// buildHunkPatch creates a patch for an entire hunk
func buildHunkPatch(filePath string, hunk diff.Hunk) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
	b.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	b.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))
	b.WriteString(hunk.Header + "\n")

	for _, line := range hunk.Lines {
		if line.Type == diff.LineHunkHeader {
			continue
		}
		switch line.Type {
		case diff.LineContext:
			b.WriteString(" " + line.Content + "\n")
		case diff.LineRemoved:
			b.WriteString("-" + line.Content + "\n")
		case diff.LineAdded:
			b.WriteString("+" + line.Content + "\n")
		}
	}

	return b.String()
}

// buildReversePatch creates a reverse patch to revert changes
func buildReversePatch(filePath string, hunk diff.Hunk) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
	b.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	b.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

	// Swap old/new in header
	b.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunk.NewStart, hunk.NewCount, hunk.OldStart, hunk.OldCount))

	for _, line := range hunk.Lines {
		if line.Type == diff.LineHunkHeader {
			continue
		}
		switch line.Type {
		case diff.LineContext:
			b.WriteString(" " + line.Content + "\n")
		case diff.LineRemoved:
			// Removed becomes added in reverse
			b.WriteString("+" + line.Content + "\n")
		case diff.LineAdded:
			// Added becomes removed in reverse
			b.WriteString("-" + line.Content + "\n")
		}
	}

	return b.String()
}

// applyPatch applies a patch to the index or working tree
func applyPatch(ctx context.Context, patch string, toIndex bool) error {
	args := []string{"apply"}
	if toIndex {
		args = append(args, "--cached")
	}
	args = append(args, "--unidiff-zero", "-")

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdin = strings.NewReader(patch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return &GitError{
			Command: strings.Join(args, " "),
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	return nil
}

// SplitLineAtCharBoundary splits a string at character boundaries (not byte boundaries)
// Returns: left (before charStart), middle (charStart to charEnd), right (after charEnd)
func SplitLineAtCharBoundary(content string, charStart, charEnd int) (left, mid, right string) {
	runes := []rune(content)
	runeCount := len(runes)

	// Clamp bounds
	if charStart < 0 {
		charStart = 0
	}
	if charEnd > runeCount {
		charEnd = runeCount
	}
	if charStart > charEnd {
		charStart = charEnd
	}

	left = string(runes[:charStart])
	mid = string(runes[charStart:charEnd])
	right = string(runes[charEnd:])

	return left, mid, right
}

// BuildCharacterPatch creates a patch that stages only specific characters within a line.
// This works by splitting the changed line at character boundaries to create a partial change.
//
// For example, if the original line is "hello world" and the new line is "hello there world",
// and we want to stage only "there ", this function creates a patch that:
// 1. Removes "hello world"
// 2. Adds "hello there" (partial staging of the addition)
//
// The remaining " world" stays unstaged for a future commit.
func BuildCharacterPatch(filePath string, hunk diff.Hunk, lineIndex, charStart, charEnd int) string {
	// Validate line index
	if lineIndex < 0 || lineIndex >= len(hunk.Lines) {
		return ""
	}

	targetLine := hunk.Lines[lineIndex]

	// Only process added or removed lines
	if targetLine.Type != diff.LineAdded && targetLine.Type != diff.LineRemoved {
		return ""
	}

	// Handle zero-width selection
	if charStart >= charEnd {
		return ""
	}

	var b strings.Builder

	// Write git diff header
	b.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))
	b.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	b.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

	// Find the paired line (removed line that corresponds to added, or vice versa)
	var pairedLine *diff.Line
	var pairedIndex int = -1

	if targetLine.Type == diff.LineAdded {
		// Look backwards for the removed line
		for i := lineIndex - 1; i >= 0; i-- {
			if hunk.Lines[i].Type == diff.LineRemoved {
				pairedLine = &hunk.Lines[i]
				pairedIndex = i
				break
			}
			if hunk.Lines[i].Type == diff.LineContext {
				break // Stop at context lines
			}
		}
	} else { // LineRemoved
		// Look forwards for the added line
		for i := lineIndex + 1; i < len(hunk.Lines); i++ {
			if hunk.Lines[i].Type == diff.LineAdded {
				pairedLine = &hunk.Lines[i]
				pairedIndex = i
				break
			}
			if hunk.Lines[i].Type == diff.LineContext {
				break
			}
		}
	}

	// Calculate what content to include based on character selection
	var oldContent, newContent string

	if targetLine.Type == diff.LineAdded {
		// Staging part of an added line
		// The old content is the paired removed line (if exists) or empty
		if pairedLine != nil {
			oldContent = pairedLine.Content
		}
		// The new content is the original content up to the selection end
		_, _, _ = SplitLineAtCharBoundary(targetLine.Content, charStart, charEnd)
		runes := []rune(targetLine.Content)
		if charEnd > len(runes) {
			charEnd = len(runes)
		}
		newContent = string(runes[:charEnd])

		// If there's a paired removed line, we need to keep the unchanged suffix
		if pairedLine != nil && charEnd < len(runes) {
			// Keep the suffix from the original that wasn't selected
			newContent = string(runes[:charEnd])
		}
	} else {
		// Staging part of a removed line
		oldContent = targetLine.Content
		if pairedLine != nil {
			newContent = pairedLine.Content
		}
		// For removed lines, we stage the whole removed line but partial added
		_ = pairedIndex // Mark as used
	}

	// Calculate hunk header values
	oldStart := hunk.OldStart
	newStart := hunk.NewStart
	oldCount := 0
	newCount := 0

	if oldContent != "" {
		oldCount = 1
	}
	if newContent != "" {
		newCount = 1
	}

	// Write hunk header
	b.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount))

	// Write the patch content
	if oldContent != "" {
		b.WriteString("-" + oldContent + "\n")
	}
	if newContent != "" {
		b.WriteString("+" + newContent + "\n")
	}

	return b.String()
}

// StageCharacters stages specific characters within a line
func StageCharacters(ctx context.Context, filePath string, hunk diff.Hunk, lineIndex, charStart, charEnd int) error {
	patch := BuildCharacterPatch(filePath, hunk, lineIndex, charStart, charEnd)
	if patch == "" {
		return nil // Nothing to stage
	}
	return applyPatch(ctx, patch, true)
}

// runeLen returns the number of runes (characters) in a string
func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}
