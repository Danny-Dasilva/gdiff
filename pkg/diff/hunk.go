package diff

// FileStatus represents the status of a file in git
type FileStatus int

const (
	StatusUnmodified FileStatus = iota
	StatusModified
	StatusAdded
	StatusDeleted
	StatusRenamed
	StatusCopied
	StatusUntracked
	StatusIgnored
	StatusUnmerged
)

func (s FileStatus) String() string {
	switch s {
	case StatusModified:
		return "M"
	case StatusAdded:
		return "A"
	case StatusDeleted:
		return "D"
	case StatusRenamed:
		return "R"
	case StatusCopied:
		return "C"
	case StatusUntracked:
		return "?"
	case StatusIgnored:
		return "!"
	case StatusUnmerged:
		return "U"
	default:
		return " "
	}
}

// FileEntry represents a file in the git status output
type FileEntry struct {
	Path       string
	OldPath    string // For renames
	Status     FileStatus
	Staged     bool
	IndexStatus FileStatus // Status in index
	WorkStatus  FileStatus // Status in working tree
}

// Line represents a single line in a diff
type Line struct {
	Type    LineType
	Content string
	OldNum  int // Line number in old file (0 if added)
	NewNum  int // Line number in new file (0 if removed)
}

// LineType indicates whether a line was added, removed, or context
type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineRemoved
	LineHunkHeader
)

// Hunk represents a diff hunk
type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string
	Lines    []Line
}

// FileDiff represents the diff for a single file
type FileDiff struct {
	OldPath string
	NewPath string
	Hunks   []Hunk
	IsBinary bool
}
