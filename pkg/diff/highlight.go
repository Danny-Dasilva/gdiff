package diff

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// CharChange represents a character-level change within a line.
// Start and End are rune (character) indices, not byte offsets.
type CharChange struct {
	Start int  // Start rune index in string
	End   int  // End rune index in string (exclusive)
	Added bool // true if added, false if removed
}

// byteToRuneOffset converts a byte offset to a rune offset within a string.
func byteToRuneOffset(s string, byteOff int) int {
	if byteOff <= 0 {
		return 0
	}
	if byteOff >= len(s) {
		return utf8.RuneCountInString(s)
	}
	return utf8.RuneCountInString(s[:byteOff])
}

// byteChangesToRuneChanges converts CharChange slices from byte offsets to rune offsets.
func byteChangesToRuneChanges(changes []CharChange, s string) []CharChange {
	if len(changes) == 0 {
		return changes
	}
	result := make([]CharChange, len(changes))
	for i, c := range changes {
		result[i] = CharChange{
			Start: byteToRuneOffset(s, c.Start),
			End:   byteToRuneOffset(s, c.End),
			Added: c.Added,
		}
	}
	return result
}

// codeTokenPattern splits on word boundaries for code-aware tokenization.
// Keeps identifiers, numbers, multi-char operators, single non-space chars,
// and whitespace runs as atomic tokens.
var codeTokenPattern = regexp.MustCompile(
	`[a-zA-Z_]\w*|[0-9]+(?:\.[0-9]+)?|&&|\|\||[<>=!]=|:=|->|\.\.\.?|[^\s]|\s+`,
)

// sentinel is a separator unlikely to appear in code tokens.
const sentinel = "\x00"

// tokenize splits a string into code-aware tokens, returning tokens and their
// byte offsets in the original string.
func tokenize(s string) (tokens []string, offsets []int) {
	locs := codeTokenPattern.FindAllStringIndex(s, -1)
	for _, loc := range locs {
		tokens = append(tokens, s[loc[0]:loc[1]])
		offsets = append(offsets, loc[0])
	}
	return
}

// ComputeCharDiff computes character-level differences between two strings
// using a word+character hybrid algorithm backed by Myers diff (sergi/go-diff).
func ComputeCharDiff(oldStr, newStr string) ([]CharChange, []CharChange) {
	if oldStr == newStr {
		return nil, nil
	}

	// Memory guard: skip diffing for extremely long strings
	if len(oldStr)+len(newStr) > 10_000 {
		return nil, nil
	}

	oldTokens, oldOffsets := tokenize(oldStr)
	newTokens, newOffsets := tokenize(newStr)

	// Handle edge cases: empty token lists
	if len(oldTokens) == 0 && len(newTokens) == 0 {
		return nil, nil
	}
	if len(oldTokens) == 0 {
		return nil, []CharChange{{Start: 0, End: utf8.RuneCountInString(newStr), Added: true}}
	}
	if len(newTokens) == 0 {
		return []CharChange{{Start: 0, End: utf8.RuneCountInString(oldStr), Added: false}}, nil
	}

	// Diff at the token level using Myers via sergi/go-diff.
	// Join tokens with a sentinel separator so diffmatchpatch treats each
	// token atomically.
	oldJoined := strings.Join(oldTokens, sentinel)
	newJoined := strings.Join(newTokens, sentinel)

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldJoined, newJoined, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	// Walk through the token-level diffs and collect changed regions.
	type region struct{ start, end int }
	var oldRegions []region
	var newRegions []region

	oldTokIdx := 0
	newTokIdx := 0

	for _, d := range diffs {
		if d.Text == "" {
			continue
		}
		// Count tokens in this diff segment by splitting on sentinel.
		parts := strings.Split(d.Text, sentinel)
		tokCount := 0
		for _, p := range parts {
			if p != "" {
				tokCount++
			}
		}

		switch d.Type {
		case diffmatchpatch.DiffEqual:
			oldTokIdx += tokCount
			newTokIdx += tokCount
		case diffmatchpatch.DiffDelete:
			if tokCount > 0 && oldTokIdx < len(oldOffsets) {
				startByte := oldOffsets[oldTokIdx]
				endIdx := oldTokIdx + tokCount - 1
				if endIdx >= len(oldTokens) {
					endIdx = len(oldTokens) - 1
				}
				endByte := oldOffsets[endIdx] + len(oldTokens[endIdx])
				oldRegions = append(oldRegions, region{startByte, endByte})
				oldTokIdx += tokCount
			}
		case diffmatchpatch.DiffInsert:
			if tokCount > 0 && newTokIdx < len(newOffsets) {
				startByte := newOffsets[newTokIdx]
				endIdx := newTokIdx + tokCount - 1
				if endIdx >= len(newTokens) {
					endIdx = len(newTokens) - 1
				}
				endByte := newOffsets[endIdx] + len(newTokens[endIdx])
				newRegions = append(newRegions, region{startByte, endByte})
				newTokIdx += tokCount
			}
		}
	}

	// Pair up delete/insert regions and refine with character-level Myers diff.
	var oldChanges []CharChange
	var newChanges []CharChange

	oi, ni := 0, 0
	for oi < len(oldRegions) || ni < len(newRegions) {
		if oi < len(oldRegions) && ni < len(newRegions) {
			// Paired region: refine with char-level diff
			oldSub := oldStr[oldRegions[oi].start:oldRegions[oi].end]
			newSub := newStr[newRegions[ni].start:newRegions[ni].end]
			subOld, subNew := charDiffMyers(
				oldSub, newSub,
				oldRegions[oi].start, newRegions[ni].start,
			)
			oldChanges = append(oldChanges, subOld...)
			newChanges = append(newChanges, subNew...)
			oi++
			ni++
		} else if oi < len(oldRegions) {
			oldChanges = append(oldChanges, CharChange{
				Start: oldRegions[oi].start,
				End:   oldRegions[oi].end,
				Added: false,
			})
			oi++
		} else {
			newChanges = append(newChanges, CharChange{
				Start: newRegions[ni].start,
				End:   newRegions[ni].end,
				Added: true,
			})
			ni++
		}
	}

	// Convert byte offsets to rune offsets for correct multibyte character handling.
	// sergi/go-diff operates on bytes, but the renderer indexes into []rune.
	return byteChangesToRuneChanges(oldChanges, oldStr), byteChangesToRuneChanges(newChanges, newStr)
}

// charDiffMyers runs character-level Myers diff on two strings and returns
// CharChange slices with byte offsets adjusted by the given base offsets.
func charDiffMyers(oldStr, newStr string, oldBase, newBase int) ([]CharChange, []CharChange) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldStr, newStr, false)

	var oldChanges []CharChange
	var newChanges []CharChange

	oldPos := oldBase
	newPos := newBase

	for _, d := range diffs {
		textLen := len(d.Text)
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			oldPos += textLen
			newPos += textLen
		case diffmatchpatch.DiffDelete:
			oldChanges = append(oldChanges, CharChange{
				Start: oldPos,
				End:   oldPos + textLen,
				Added: false,
			})
			oldPos += textLen
		case diffmatchpatch.DiffInsert:
			newChanges = append(newChanges, CharChange{
				Start: newPos,
				End:   newPos + textLen,
				Added: true,
			})
			newPos += textLen
		}
	}

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

// levenshteinSimilarity computes normalized similarity between two strings.
// Returns a value between 0.0 (completely different) and 1.0 (identical).
func levenshteinSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	ar := []rune(a)
	br := []rune(b)
	la, lb := len(ar), len(br)
	if la == 0 && lb == 0 {
		return 1.0
	}
	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}

	// Levenshtein distance using two-row approach (O(min(m,n)) space)
	if la < lb {
		ar, br = br, ar
		la, lb = lb, la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}
	dist := prev[lb]
	return 1.0 - float64(dist)/float64(maxLen)
}

// PairAdjacentLines finds pairs of removed/added lines that are likely related
// using Levenshtein similarity scoring for better matching.
// Returns pairs of (oldIndex, newIndex) for lines that should be compared.
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

		removeCount := removeEnd - removeStart
		addCount := addEnd - addStart

		if removeCount == 0 || addCount == 0 {
			continue
		}

		// Fast path: if counts are equal, check if sequential pairing is good
		if removeCount == addCount {
			totalSim := 0.0
			for j := 0; j < removeCount; j++ {
				totalSim += levenshteinSimilarity(
					lines[removeStart+j].Content,
					lines[addStart+j].Content,
				)
			}
			avgSim := totalSim / float64(removeCount)
			if avgSim > 0.5 {
				for j := 0; j < removeCount; j++ {
					pairs = append(pairs, [2]int{removeStart + j, addStart + j})
				}
				continue
			}
		}

		// Build similarity matrix and do greedy matching
		usedAdd := make([]bool, addCount)

		for ri := 0; ri < removeCount; ri++ {
			bestJ := -1
			bestSim := 0.4 // threshold
			for aj := 0; aj < addCount; aj++ {
				if usedAdd[aj] {
					continue
				}
				sim := levenshteinSimilarity(
					lines[removeStart+ri].Content,
					lines[addStart+aj].Content,
				)
				if sim > bestSim {
					bestSim = sim
					bestJ = aj
				}
			}
			if bestJ >= 0 {
				pairs = append(pairs, [2]int{removeStart + ri, addStart + bestJ})
				usedAdd[bestJ] = true
			}
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
