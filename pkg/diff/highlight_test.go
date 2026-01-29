package diff

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestComputeCharDiff(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		new         string
		wantOldDiff bool // expect old changes to be non-empty
		wantNewDiff bool // expect new changes to be non-empty
	}{
		{
			name:        "simple change",
			old:         "hello world",
			new:         "hello there",
			wantOldDiff: true,
			wantNewDiff: true,
		},
		{
			name:        "prefix change",
			old:         "foo bar",
			new:         "baz bar",
			wantOldDiff: true,
			wantNewDiff: true,
		},
		{
			name:        "identical strings",
			old:         "same",
			new:         "same",
			wantOldDiff: false,
			wantNewDiff: false,
		},
		{
			name:        "empty to content",
			old:         "",
			new:         "added",
			wantOldDiff: false,
			wantNewDiff: true,
		},
		{
			name:        "content to empty",
			old:         "removed",
			new:         "",
			wantOldDiff: true,
			wantNewDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldChanges, newChanges := ComputeCharDiff(tt.old, tt.new)
			hasOld := len(oldChanges) > 0
			hasNew := len(newChanges) > 0
			if hasOld != tt.wantOldDiff {
				t.Errorf("old changes: got non-empty=%v, want non-empty=%v (len=%d)", hasOld, tt.wantOldDiff, len(oldChanges))
			}
			if hasNew != tt.wantNewDiff {
				t.Errorf("new changes: got non-empty=%v, want non-empty=%v (len=%d)", hasNew, tt.wantNewDiff, len(newChanges))
			}
		})
	}
}

func TestPairAdjacentLines(t *testing.T) {
	// Use realistic code content so Levenshtein similarity exceeds the 0.4 threshold
	lines := []Line{
		{Type: LineContext, Content: "func main() {"},
		{Type: LineRemoved, Content: "    value := computeOld(input)"},
		{Type: LineRemoved, Content: "    result := processOld(value)"},
		{Type: LineAdded, Content: "    value := computeNew(input)"},
		{Type: LineAdded, Content: "    result := processNew(value)"},
		{Type: LineContext, Content: "    return result"},
		{Type: LineRemoved, Content: "    fmt.Println(oldMsg)"},
		{Type: LineAdded, Content: "    fmt.Println(newMsg)"},
	}

	pairs := PairAdjacentLines(lines)

	// Should pair: (1,3), (2,4), (6,7)
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d: %v", len(pairs), pairs)
	}

	// First pair
	if pairs[0][0] != 1 || pairs[0][1] != 3 {
		t.Errorf("expected pair (1,3), got (%d,%d)", pairs[0][0], pairs[0][1])
	}

	// Second pair
	if pairs[1][0] != 2 || pairs[1][1] != 4 {
		t.Errorf("expected pair (2,4), got (%d,%d)", pairs[1][0], pairs[1][1])
	}

	// Third pair
	if pairs[2][0] != 6 || pairs[2][1] != 7 {
		t.Errorf("expected pair (6,7), got (%d,%d)", pairs[2][0], pairs[2][1])
	}
}

func TestIndicesToRanges(t *testing.T) {
	tests := []struct {
		name    string
		indices []int
		want    int
	}{
		{
			name:    "single index",
			indices: []int{5},
			want:    1,
		},
		{
			name:    "contiguous",
			indices: []int{1, 2, 3, 4},
			want:    1,
		},
		{
			name:    "two ranges",
			indices: []int{1, 2, 5, 6},
			want:    2,
		},
		{
			name:    "scattered",
			indices: []int{1, 3, 5},
			want:    3,
		},
		{
			name:    "empty",
			indices: []int{},
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges := indicesToRanges(tt.indices, true)
			if len(ranges) != tt.want {
				t.Errorf("got %d ranges, want %d", len(ranges), tt.want)
			}
		})
	}
}

// --- Behavioral tests ---

// validateCharChanges checks that all CharChange ranges are valid for the given string length.
func validateCharChanges(t *testing.T, label string, changes []CharChange, strLen int) {
	t.Helper()
	for i, c := range changes {
		if c.Start < 0 {
			t.Errorf("%s[%d]: Start %d < 0", label, i, c.Start)
		}
		if c.End <= c.Start {
			t.Errorf("%s[%d]: End %d <= Start %d", label, i, c.End, c.Start)
		}
		if c.End > strLen {
			t.Errorf("%s[%d]: End %d > string length %d", label, i, c.End, strLen)
		}
		if i > 0 && c.Start < changes[i-1].End {
			t.Errorf("%s[%d]: Start %d overlaps previous End %d", label, i, c.Start, changes[i-1].End)
		}
	}
}

func TestComputeCharDiffUnicode(t *testing.T) {
	tests := []struct {
		name string
		old  string
		new  string
	}{
		{
			name: "accent change",
			old:  "h\u00e9llo",
			new:  "h\u00ebllo",
		},
		{
			name: "emoji change",
			old:  "hello \U0001f30d",
			new:  "hello \U0001f30e",
		},
		{
			name: "CJK change",
			old:  "\u65e5\u672c\u8a9e",
			new:  "\u65e5\u672c\u4eba",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldChanges, newChanges := ComputeCharDiff(tt.old, tt.new)

			if tt.old != tt.new {
				if len(oldChanges) == 0 && len(newChanges) == 0 {
					t.Error("expected at least one side to have changes for different strings")
				}
			}

			validateCharChanges(t, "old", oldChanges, utf8.RuneCountInString(tt.old))
			validateCharChanges(t, "new", newChanges, utf8.RuneCountInString(tt.new))
		})
	}
}

func TestComputeCharDiffLongLines(t *testing.T) {
	big := strings.Repeat("a", 6000)

	t.Run("long similar strings", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff(big, big+"x")
		// Either both nil (guarded) or valid results
		if oldChanges != nil {
			validateCharChanges(t, "old", oldChanges, len(big))
		}
		if newChanges != nil {
			validateCharChanges(t, "new", newChanges, len(big+"x"))
		}
	})

	t.Run("very long identical strings", func(t *testing.T) {
		veryBig := strings.Repeat("x", 15000)
		oldChanges, newChanges := ComputeCharDiff(veryBig, veryBig)
		if len(oldChanges) != 0 || len(newChanges) != 0 {
			t.Errorf("identical long strings should have no changes, got old=%d new=%d",
				len(oldChanges), len(newChanges))
		}
	})
}

func TestComputeCharDiffWordBoundary(t *testing.T) {
	t.Run("return err vs return nil", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("return err", "return nil")

		if len(oldChanges) == 0 {
			t.Fatal("expected old changes for 'return err' vs 'return nil'")
		}
		if len(newChanges) == 0 {
			t.Fatal("expected new changes for 'return err' vs 'return nil'")
		}

		validateCharChanges(t, "old", oldChanges, len("return err"))
		validateCharChanges(t, "new", newChanges, len("return nil"))

		// Changes should not be in the shared prefix "return "
		for _, c := range oldChanges {
			if c.End <= 7 {
				t.Errorf("old change [%d,%d) is in shared prefix 'return '", c.Start, c.End)
			}
		}
		for _, c := range newChanges {
			if c.End <= 7 {
				t.Errorf("new change [%d,%d) is in shared prefix 'return '", c.Start, c.End)
			}
		}
	})

	t.Run("single word change in sentence", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("the quick brown fox", "the quick red fox")

		if len(oldChanges) == 0 || len(newChanges) == 0 {
			t.Fatal("expected changes on both sides")
		}

		validateCharChanges(t, "old", oldChanges, len("the quick brown fox"))
		validateCharChanges(t, "new", newChanges, len("the quick red fox"))
	})
}

func TestComputeCharDiffEdgeCases(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("", "")
		if len(oldChanges) != 0 || len(newChanges) != 0 {
			t.Errorf("expected no changes for empty strings")
		}
	})

	t.Run("completely different strings", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("abc", "xyz")
		if len(oldChanges) == 0 || len(newChanges) == 0 {
			t.Error("expected changes for completely different strings")
		}
		validateCharChanges(t, "old", oldChanges, 3)
		validateCharChanges(t, "new", newChanges, 3)
	})

	t.Run("whitespace only change", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("hello world", "hello  world")
		if len(oldChanges) == 0 && len(newChanges) == 0 {
			t.Error("expected changes for whitespace difference")
		}
	})

	t.Run("single char difference", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("cat", "bat")
		if len(oldChanges) == 0 || len(newChanges) == 0 {
			t.Error("expected changes for single char difference")
		}
		validateCharChanges(t, "old", oldChanges, 3)
		validateCharChanges(t, "new", newChanges, 3)
	})

	t.Run("repeated chars - insertion", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("aaa", "aaaa")
		// The word-level tokenizer treats "aaa" and "aaaa" as single tokens,
		// so DiffCleanupSemantic may merge them. Either non-empty changes
		// or nil (if cleanup decides it's a single token substitution) is acceptable.
		if oldChanges != nil {
			validateCharChanges(t, "old", oldChanges, 3)
		}
		if newChanges != nil {
			validateCharChanges(t, "new", newChanges, 4)
		}
	})
}

func TestPairAdjacentLinesUnequal(t *testing.T) {
	t.Run("more removes than adds", func(t *testing.T) {
		// Use similar content so Levenshtein similarity > 0.4 threshold
		lines := []Line{
			{Type: LineRemoved, Content: "return handleError(err)"},
			{Type: LineRemoved, Content: "return fmt.Errorf(msg)"},
			{Type: LineRemoved, Content: "return nil"},
			{Type: LineAdded, Content: "return handleError(nil)"},
		}
		pairs := PairAdjacentLines(lines)
		if len(pairs) < 1 {
			t.Fatal("expected at least 1 pair")
		}
	})

	t.Run("more adds than removes", func(t *testing.T) {
		lines := []Line{
			{Type: LineRemoved, Content: "func processData(input string) error {"},
			{Type: LineAdded, Content: "func processData(input string) (string, error) {"},
			{Type: LineAdded, Content: "    result := transform(input)"},
			{Type: LineAdded, Content: "    return result, nil"},
		}
		pairs := PairAdjacentLines(lines)
		if len(pairs) < 1 {
			t.Fatal("expected at least 1 pair")
		}
	})

	t.Run("single remove no adds", func(t *testing.T) {
		lines := []Line{
			{Type: LineRemoved, Content: "old1"},
			{Type: LineContext, Content: "ctx"},
		}
		pairs := PairAdjacentLines(lines)
		if len(pairs) != 0 {
			t.Errorf("expected 0 pairs for remove-only, got %d", len(pairs))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		pairs := PairAdjacentLines([]Line{})
		if len(pairs) != 0 {
			t.Errorf("expected 0 pairs for empty input, got %d", len(pairs))
		}
	})
}

func TestCharChangeRangesValid(t *testing.T) {
	inputs := []struct {
		name string
		old  string
		new  string
	}{
		{"simple substitution", "hello", "world"},
		{"prefix shared", "foobar", "foobaz"},
		{"suffix shared", "abcdef", "xyzdef"},
		{"middle change", "abcdef", "abXXef"},
		{"insertion", "abc", "aXbc"},
		{"deletion", "abcd", "acd"},
		{"empty old", "", "something"},
		{"empty new", "something", ""},
		{"both empty", "", ""},
		{"identical", "same", "same"},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			oldChanges, newChanges := ComputeCharDiff(tt.old, tt.new)

			validateCharChanges(t, "old", oldChanges, utf8.RuneCountInString(tt.old))
			validateCharChanges(t, "new", newChanges, utf8.RuneCountInString(tt.new))

			// Verify Added flag consistency
			for i, c := range oldChanges {
				if c.Added {
					t.Errorf("old change[%d] has Added=true, expected false", i)
				}
			}
			for i, c := range newChanges {
				if !c.Added {
					t.Errorf("new change[%d] has Added=false, expected true", i)
				}
			}

			// Verify ranges are sorted by Start
			for i := 1; i < len(oldChanges); i++ {
				if oldChanges[i].Start <= oldChanges[i-1].Start {
					t.Errorf("old changes not sorted: [%d].Start=%d <= [%d].Start=%d",
						i, oldChanges[i].Start, i-1, oldChanges[i-1].Start)
				}
			}
			for i := 1; i < len(newChanges); i++ {
				if newChanges[i].Start <= newChanges[i-1].Start {
					t.Errorf("new changes not sorted: [%d].Start=%d <= [%d].Start=%d",
						i, newChanges[i].Start, i-1, newChanges[i-1].Start)
				}
			}
		})
	}
}

func TestComputeCharDiffRuneOffsets(t *testing.T) {
	// Verify that CharChange offsets are rune-based, not byte-based.
	// "héllo" is 6 bytes but 5 runes; "hëllo" is also 6 bytes, 5 runes.
	// The changed rune is at rune index 1 (byte index 1-2).
	t.Run("multibyte rune offset", func(t *testing.T) {
		oldChanges, newChanges := ComputeCharDiff("héllo", "hëllo")

		if len(oldChanges) == 0 || len(newChanges) == 0 {
			t.Fatal("expected changes for accent swap")
		}

		// Offsets must be rune indices, not byte indices.
		// The string is 5 runes long; no offset should exceed 5.
		runeLen := utf8.RuneCountInString("héllo") // 5
		for _, c := range oldChanges {
			if c.End > runeLen {
				t.Errorf("old change end %d exceeds rune count %d (would be valid as byte offset but not rune)", c.End, runeLen)
			}
		}
		for _, c := range newChanges {
			if c.End > runeLen {
				t.Errorf("new change end %d exceeds rune count %d", c.End, runeLen)
			}
		}
	})

	t.Run("CJK rune offsets", func(t *testing.T) {
		// 日本語 = 3 runes, 9 bytes; 日本人 = 3 runes, 9 bytes
		oldChanges, newChanges := ComputeCharDiff("日本語", "日本人")

		if len(oldChanges) == 0 || len(newChanges) == 0 {
			t.Fatal("expected changes for CJK")
		}

		// All offsets must be in [0, 3] range (rune count)
		for _, c := range oldChanges {
			if c.Start < 0 || c.End > 3 {
				t.Errorf("old change [%d,%d) out of rune range [0,3]", c.Start, c.End)
			}
		}
		for _, c := range newChanges {
			if c.Start < 0 || c.End > 3 {
				t.Errorf("new change [%d,%d) out of rune range [0,3]", c.Start, c.End)
			}
		}
	})

	t.Run("emoji rune offsets", func(t *testing.T) {
		// "hi 🌍" = 4 runes (7 bytes); "hi 🌎" = 4 runes (7 bytes)
		oldChanges, newChanges := ComputeCharDiff("hi 🌍", "hi 🌎")

		if len(oldChanges) == 0 || len(newChanges) == 0 {
			t.Fatal("expected changes for emoji swap")
		}

		runeLen := utf8.RuneCountInString("hi 🌍") // 4
		for _, c := range oldChanges {
			if c.End > runeLen {
				t.Errorf("old change end %d exceeds rune count %d", c.End, runeLen)
			}
		}
		for _, c := range newChanges {
			if c.End > runeLen {
				t.Errorf("new change end %d exceeds rune count %d", c.End, runeLen)
			}
		}
	})
}

func TestComputeHighlightedDiffIntegration(t *testing.T) {
	t.Run("mixed hunk", func(t *testing.T) {
		hunk := Hunk{
			Lines: []Line{
				{Type: LineContext, Content: "unchanged line"},
				{Type: LineRemoved, Content: "old value = 42"},
				{Type: LineAdded, Content: "new value = 99"},
				{Type: LineContext, Content: "another unchanged"},
			},
		}

		result := ComputeHighlightedDiff(hunk)

		if len(result) != len(hunk.Lines) {
			t.Fatalf("expected %d highlighted lines, got %d", len(hunk.Lines), len(result))
		}

		// Context lines should have no changes
		if len(result[0].Changes) != 0 {
			t.Errorf("context line 0 should have no changes, got %d", len(result[0].Changes))
		}
		if len(result[3].Changes) != 0 {
			t.Errorf("context line 3 should have no changes, got %d", len(result[3].Changes))
		}

		// Paired remove/add lines should have changes
		if len(result[1].Changes) == 0 {
			t.Error("removed line should have character-level changes")
		}
		if len(result[2].Changes) == 0 {
			t.Error("added line should have character-level changes")
		}
	})

	t.Run("empty hunk", func(t *testing.T) {
		hunk := Hunk{Lines: []Line{}}
		result := ComputeHighlightedDiff(hunk)
		if len(result) != 0 {
			t.Errorf("empty hunk should produce 0 lines, got %d", len(result))
		}
	})

	t.Run("multiple paired groups", func(t *testing.T) {
		hunk := Hunk{
			Lines: []Line{
				{Type: LineRemoved, Content: "func foo() {"},
				{Type: LineAdded, Content: "func bar() {"},
				{Type: LineContext, Content: "    // body"},
				{Type: LineRemoved, Content: "    return 1"},
				{Type: LineAdded, Content: "    return 2"},
			},
		}

		result := ComputeHighlightedDiff(hunk)
		if len(result) != 5 {
			t.Fatalf("expected 5 lines, got %d", len(result))
		}

		if len(result[0].Changes) == 0 {
			t.Error("first removed line should have changes")
		}
		if len(result[1].Changes) == 0 {
			t.Error("first added line should have changes")
		}
		if len(result[2].Changes) != 0 {
			t.Error("context line should have no changes")
		}
		if len(result[3].Changes) == 0 {
			t.Error("second removed line should have changes")
		}
		if len(result[4].Changes) == 0 {
			t.Error("second added line should have changes")
		}
	})
}
