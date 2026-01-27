package diff

import (
	"testing"
)

func TestParse(t *testing.T) {
	diffOutput := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main

-import "fmt"
+import (
+	"fmt"
+)

 func main() {`

	result := Parse(diffOutput)

	if len(result) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(result))
	}

	fd := result[0]
	if fd.OldPath != "main.go" {
		t.Errorf("expected OldPath 'main.go', got '%s'", fd.OldPath)
	}
	if fd.NewPath != "main.go" {
		t.Errorf("expected NewPath 'main.go', got '%s'", fd.NewPath)
	}

	if len(fd.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(fd.Hunks))
	}

	hunk := fd.Hunks[0]
	if hunk.OldStart != 1 || hunk.OldCount != 5 {
		t.Errorf("expected old range 1,5, got %d,%d", hunk.OldStart, hunk.OldCount)
	}
	if hunk.NewStart != 1 || hunk.NewCount != 6 {
		t.Errorf("expected new range 1,6, got %d,%d", hunk.NewStart, hunk.NewCount)
	}

	// Count line types
	var context, added, removed int
	for _, line := range hunk.Lines {
		switch line.Type {
		case LineContext:
			context++
		case LineAdded:
			added++
		case LineRemoved:
			removed++
		}
	}

	if removed != 1 {
		t.Errorf("expected 1 removed line, got %d", removed)
	}
	if added != 3 {
		t.Errorf("expected 3 added lines, got %d", added)
	}
}

func TestParseBinaryFile(t *testing.T) {
	diffOutput := `diff --git a/image.png b/image.png
Binary files a/image.png and b/image.png differ`

	result := Parse(diffOutput)

	if len(result) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(result))
	}

	if !result[0].IsBinary {
		t.Error("expected binary file to be marked as binary")
	}
}

func TestParseMultipleFiles(t *testing.T) {
	diffOutput := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package foo
+
 func A() {}
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1 +1 @@
-old
+new`

	result := Parse(diffOutput)

	if len(result) != 2 {
		t.Fatalf("expected 2 file diffs, got %d", len(result))
	}

	if result[0].OldPath != "file1.go" {
		t.Errorf("expected first file 'file1.go', got '%s'", result[0].OldPath)
	}
	if result[1].OldPath != "file2.go" {
		t.Errorf("expected second file 'file2.go', got '%s'", result[1].OldPath)
	}
}
