package diff

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)$`)
	fileHeaderRe = regexp.MustCompile(`^diff --git a/(.*) b/(.*)$`)
)

// Parse parses a unified diff output into FileDiffs
func Parse(diffOutput string) []FileDiff {
	var result []FileDiff
	lines := strings.Split(diffOutput, "\n")

	var currentFile *FileDiff
	var currentHunk *Hunk
	oldLineNum, newLineNum := 0, 0

	for _, line := range lines {

		// Check for file header
		if matches := fileHeaderRe.FindStringSubmatch(line); matches != nil {
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				}
				result = append(result, *currentFile)
			}
			currentFile = &FileDiff{
				OldPath: matches[1],
				NewPath: matches[2],
			}
			currentHunk = nil
			continue
		}

		// Check for binary file
		if strings.HasPrefix(line, "Binary files") && currentFile != nil {
			currentFile.IsBinary = true
			continue
		}

		// Check for hunk header
		if matches := hunkHeaderRe.FindStringSubmatch(line); matches != nil {
			if currentHunk != nil && currentFile != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			oldStart, _ := strconv.Atoi(matches[1])
			oldCount := 1
			if matches[2] != "" {
				oldCount, _ = strconv.Atoi(matches[2])
			}
			newStart, _ := strconv.Atoi(matches[3])
			newCount := 1
			if matches[4] != "" {
				newCount, _ = strconv.Atoi(matches[4])
			}

			currentHunk = &Hunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
				Header:   line,
			}
			currentHunk.Lines = append(currentHunk.Lines, Line{
				Type:    LineHunkHeader,
				Content: line,
			})

			oldLineNum = oldStart
			newLineNum = newStart
			continue
		}

		// Skip other header lines
		if currentHunk == nil {
			continue
		}

		// Parse diff lines
		if len(line) == 0 {
			// Empty context line
			currentHunk.Lines = append(currentHunk.Lines, Line{
				Type:    LineContext,
				Content: "",
				OldNum:  oldLineNum,
				NewNum:  newLineNum,
			})
			oldLineNum++
			newLineNum++
		} else {
			prefix := line[0]
			content := ""
			if len(line) > 1 {
				content = line[1:]
			}

			switch prefix {
			case '+':
				currentHunk.Lines = append(currentHunk.Lines, Line{
					Type:    LineAdded,
					Content: content,
					NewNum:  newLineNum,
				})
				newLineNum++
			case '-':
				currentHunk.Lines = append(currentHunk.Lines, Line{
					Type:    LineRemoved,
					Content: content,
					OldNum:  oldLineNum,
				})
				oldLineNum++
			case ' ':
				currentHunk.Lines = append(currentHunk.Lines, Line{
					Type:    LineContext,
					Content: content,
					OldNum:  oldLineNum,
					NewNum:  newLineNum,
				})
				oldLineNum++
				newLineNum++
			case '\\':
				// "\ No newline at end of file" - skip
				continue
			}
		}
	}

	// Don't forget the last file/hunk
	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		result = append(result, *currentFile)
	}

	return result
}
