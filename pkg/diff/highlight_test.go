package diff

import (
	"testing"
)

func TestComputeCharDiff(t *testing.T) {
	tests := []struct {
		name       string
		old        string
		new        string
		wantOldLen int
		wantNewLen int
	}{
		{
			name:       "simple change",
			old:        "hello world",
			new:        "hello there",
			wantOldLen: 2, // LCS finds 'w' and 'orld' as separate ranges
			wantNewLen: 2, // LCS finds 'th' and 're' as separate ranges
		},
		{
			name:       "prefix change",
			old:        "foo bar",
			new:        "baz bar",
			wantOldLen: 1,
			wantNewLen: 1,
		},
		{
			name:       "identical strings",
			old:        "same",
			new:        "same",
			wantOldLen: 0,
			wantNewLen: 0,
		},
		{
			name:       "empty to content",
			old:        "",
			new:        "added",
			wantOldLen: 0,
			wantNewLen: 1,
		},
		{
			name:       "content to empty",
			old:        "removed",
			new:        "",
			wantOldLen: 1,
			wantNewLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldChanges, newChanges := ComputeCharDiff(tt.old, tt.new)
			if len(oldChanges) != tt.wantOldLen {
				t.Errorf("old changes: got %d ranges, want %d", len(oldChanges), tt.wantOldLen)
			}
			if len(newChanges) != tt.wantNewLen {
				t.Errorf("new changes: got %d ranges, want %d", len(newChanges), tt.wantNewLen)
			}
		})
	}
}

func TestPairAdjacentLines(t *testing.T) {
	lines := []Line{
		{Type: LineContext, Content: "context1"},
		{Type: LineRemoved, Content: "old1"},
		{Type: LineRemoved, Content: "old2"},
		{Type: LineAdded, Content: "new1"},
		{Type: LineAdded, Content: "new2"},
		{Type: LineContext, Content: "context2"},
		{Type: LineRemoved, Content: "old3"},
		{Type: LineAdded, Content: "new3"},
	}

	pairs := PairAdjacentLines(lines)

	// Should pair: (1,3), (2,4), (6,7)
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
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
