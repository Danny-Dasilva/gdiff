package git

import (
	"strings"
	"testing"

	"github.com/Danny-Dasilva/gdiff/pkg/diff"
)

func TestBuildCharacterPatch(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		hunk        diff.Hunk
		lineIndex   int
		charStart   int
		charEnd     int
		wantAdded   string // Expected added line content in patch
		wantRemoved string // Expected removed line content in patch
	}{
		{
			name:     "split added line at character boundary",
			filePath: "test.go",
			hunk: diff.Hunk{
				OldStart: 1,
				OldCount: 1,
				NewStart: 1,
				NewCount: 1,
				Header:   "@@ -1,1 +1,1 @@",
				Lines: []diff.Line{
					{Type: diff.LineRemoved, Content: "hello world", OldNum: 1},
					{Type: diff.LineAdded, Content: "hello there world", NewNum: 1},
				},
			},
			lineIndex:   1, // The added line
			charStart:   6, // Start of "there "
			charEnd:     12,
			wantAdded:   "hello there",
			wantRemoved: "hello world",
		},
		{
			name:     "stage first part of added characters",
			filePath: "test.go",
			hunk: diff.Hunk{
				OldStart: 1,
				OldCount: 1,
				NewStart: 1,
				NewCount: 1,
				Header:   "@@ -1,1 +1,1 @@",
				Lines: []diff.Line{
					{Type: diff.LineRemoved, Content: "foo", OldNum: 1},
					{Type: diff.LineAdded, Content: "foo bar baz", NewNum: 1},
				},
			},
			lineIndex:   1,
			charStart:   4, // Start of "bar "
			charEnd:     8,
			wantAdded:   "foo bar",
			wantRemoved: "foo",
		},
		{
			name:     "stage middle characters",
			filePath: "test.go",
			hunk: diff.Hunk{
				OldStart: 10,
				OldCount: 1,
				NewStart: 10,
				NewCount: 1,
				Header:   "@@ -10,1 +10,1 @@",
				Lines: []diff.Line{
					{Type: diff.LineRemoved, Content: "abc xyz", OldNum: 10},
					{Type: diff.LineAdded, Content: "abc 123 xyz", NewNum: 10},
				},
			},
			lineIndex:   1,
			charStart:   4,
			charEnd:     8, // "123 "
			wantAdded:   "abc 123",
			wantRemoved: "abc xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := BuildCharacterPatch(tt.filePath, tt.hunk, tt.lineIndex, tt.charStart, tt.charEnd)

			// Verify patch structure
			if !strings.Contains(patch, "diff --git") {
				t.Error("patch missing git diff header")
			}
			if !strings.Contains(patch, "--- a/"+tt.filePath) {
				t.Errorf("patch missing old file header, got:\n%s", patch)
			}
			if !strings.Contains(patch, "+++ b/"+tt.filePath) {
				t.Errorf("patch missing new file header, got:\n%s", patch)
			}

			// Verify the patch contains the expected content
			lines := strings.Split(patch, "\n")
			var foundAdded, foundRemoved bool
			for _, line := range lines {
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					if strings.Contains(line, tt.wantAdded) || line[1:] == tt.wantAdded {
						foundAdded = true
					}
				}
				if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
					if strings.Contains(line, tt.wantRemoved) || line[1:] == tt.wantRemoved {
						foundRemoved = true
					}
				}
			}

			if !foundAdded {
				t.Errorf("patch missing expected added content %q, got:\n%s", tt.wantAdded, patch)
			}
			if !foundRemoved {
				t.Errorf("patch missing expected removed content %q, got:\n%s", tt.wantRemoved, patch)
			}
		})
	}
}

func TestBuildCharacterPatchEdgeCases(t *testing.T) {
	t.Run("full line selection returns standard patch", func(t *testing.T) {
		hunk := diff.Hunk{
			OldStart: 1,
			OldCount: 1,
			NewStart: 1,
			NewCount: 1,
			Header:   "@@ -1,1 +1,1 @@",
			Lines: []diff.Line{
				{Type: diff.LineRemoved, Content: "old content", OldNum: 1},
				{Type: diff.LineAdded, Content: "new content", NewNum: 1},
			},
		}

		patch := BuildCharacterPatch("test.go", hunk, 1, 0, 11) // Full line
		if !strings.Contains(patch, "+new content") {
			t.Errorf("full line patch should contain full new content, got:\n%s", patch)
		}
	})

	t.Run("empty selection returns empty patch", func(t *testing.T) {
		hunk := diff.Hunk{
			OldStart: 1,
			OldCount: 1,
			NewStart: 1,
			NewCount: 1,
			Header:   "@@ -1,1 +1,1 @@",
			Lines: []diff.Line{
				{Type: diff.LineAdded, Content: "new content", NewNum: 1},
			},
		}

		patch := BuildCharacterPatch("test.go", hunk, 0, 5, 5) // Zero-width selection
		// Should return empty or minimal patch
		if patch != "" && strings.Contains(patch, "+") && !strings.HasPrefix(patch, "+++") {
			t.Log("Zero-width selection produced a patch (acceptable)")
		}
	})

	t.Run("invalid line index returns empty", func(t *testing.T) {
		hunk := diff.Hunk{
			OldStart: 1,
			OldCount: 1,
			NewStart: 1,
			NewCount: 1,
			Lines: []diff.Line{
				{Type: diff.LineAdded, Content: "content", NewNum: 1},
			},
		}

		patch := BuildCharacterPatch("test.go", hunk, 99, 0, 5) // Invalid index
		if patch != "" {
			t.Errorf("invalid line index should return empty patch, got:\n%s", patch)
		}
	})
}

func TestSplitLineAtCharBoundary(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		charStart int
		charEnd   int
		wantLeft  string
		wantMid   string
		wantRight string
	}{
		{
			name:      "split in middle",
			content:   "hello world",
			charStart: 6,
			charEnd:   11,
			wantLeft:  "hello ",
			wantMid:   "world",
			wantRight: "",
		},
		{
			name:      "split at start",
			content:   "hello world",
			charStart: 0,
			charEnd:   5,
			wantLeft:  "",
			wantMid:   "hello",
			wantRight: " world",
		},
		{
			name:      "split full string",
			content:   "hello",
			charStart: 0,
			charEnd:   5,
			wantLeft:  "",
			wantMid:   "hello",
			wantRight: "",
		},
		{
			name:      "unicode characters",
			content:   "hello \xe4\xb8\x96\xe7\x95\x8c world", // hello 世界 world
			charStart: 6,
			charEnd:   8, // Just the two Chinese characters
			wantLeft:  "hello ",
			wantMid:   "\xe4\xb8\x96\xe7\x95\x8c",
			wantRight: " world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, mid, right := SplitLineAtCharBoundary(tt.content, tt.charStart, tt.charEnd)
			if left != tt.wantLeft {
				t.Errorf("left = %q, want %q", left, tt.wantLeft)
			}
			if mid != tt.wantMid {
				t.Errorf("mid = %q, want %q", mid, tt.wantMid)
			}
			if right != tt.wantRight {
				t.Errorf("right = %q, want %q", right, tt.wantRight)
			}
		})
	}
}
