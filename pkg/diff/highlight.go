package diff

// CharChange represents a character-level change within a line
type CharChange struct {
	Start int  // Start index in string
	End   int  // End index in string
	Added bool // true if added, false if removed
}

// ComputeCharDiff computes character-level differences between two strings
// using a simple LCS-based algorithm
func ComputeCharDiff(oldStr, newStr string) ([]CharChange, []CharChange) {
	oldRunes := []rune(oldStr)
	newRunes := []rune(newStr)

	// Compute LCS matrix
	m, n := len(oldRunes), len(newRunes)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldRunes[i-1] == newRunes[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else {
				lcs[i][j] = max(lcs[i-1][j], lcs[i][j-1])
			}
		}
	}

	// Backtrack to find changes
	var oldChanges []CharChange
	var newChanges []CharChange

	i, j := m, n
	var oldDel, newAdd []int // Indices of changes

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldRunes[i-1] == newRunes[j-1] {
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			newAdd = append([]int{j - 1}, newAdd...)
			j--
		} else {
			oldDel = append([]int{i - 1}, oldDel...)
			i--
		}
	}

	// Convert indices to ranges
	oldChanges = indicesToRanges(oldDel, false)
	newChanges = indicesToRanges(newAdd, true)

	return oldChanges, newChanges
}

// indicesToRanges converts a list of indices to contiguous ranges
func indicesToRanges(indices []int, added bool) []CharChange {
	if len(indices) == 0 {
		return nil
	}

	var ranges []CharChange
	start := indices[0]
	end := indices[0]

	for i := 1; i < len(indices); i++ {
		if indices[i] == end+1 {
			end = indices[i]
		} else {
			ranges = append(ranges, CharChange{
				Start: start,
				End:   end + 1,
				Added: added,
			})
			start = indices[i]
			end = indices[i]
		}
	}

	ranges = append(ranges, CharChange{
		Start: start,
		End:   end + 1,
		Added: added,
	})

	return ranges
}

// PairAdjacentLines finds pairs of removed/added lines that are likely related
// Returns pairs of (oldIndex, newIndex) for lines that should be compared
func PairAdjacentLines(lines []Line) [][2]int {
	var pairs [][2]int

	i := 0
	for i < len(lines) {
		// Look for a removed line
		if lines[i].Type != LineRemoved {
			i++
			continue
		}

		// Find consecutive removed lines
		removeStart := i
		for i < len(lines) && lines[i].Type == LineRemoved {
			i++
		}
		removeEnd := i

		// Find consecutive added lines
		addStart := i
		for i < len(lines) && lines[i].Type == LineAdded {
			i++
		}
		addEnd := i

		// Pair them up (one-to-one for now)
		removeCount := removeEnd - removeStart
		addCount := addEnd - addStart

		for j := 0; j < min(removeCount, addCount); j++ {
			pairs = append(pairs, [2]int{removeStart + j, addStart + j})
		}
	}

	return pairs
}

// HighlightedLine represents a line with character-level highlighting
type HighlightedLine struct {
	Line    Line
	Changes []CharChange
}

// ComputeHighlightedDiff computes character-level highlighting for a hunk
func ComputeHighlightedDiff(hunk Hunk) []HighlightedLine {
	result := make([]HighlightedLine, len(hunk.Lines))

	// Initialize with no highlighting
	for i, line := range hunk.Lines {
		result[i] = HighlightedLine{Line: line}
	}

	// Find paired lines
	pairs := PairAdjacentLines(hunk.Lines)

	// Compute character diffs for each pair
	for _, pair := range pairs {
		oldIdx, newIdx := pair[0], pair[1]
		if oldIdx >= len(hunk.Lines) || newIdx >= len(hunk.Lines) {
			continue
		}

		oldLine := hunk.Lines[oldIdx].Content
		newLine := hunk.Lines[newIdx].Content

		oldChanges, newChanges := ComputeCharDiff(oldLine, newLine)
		result[oldIdx].Changes = oldChanges
		result[newIdx].Changes = newChanges
	}

	return result
}
