package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEditor(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "EDITOR set",
			envVars:  map[string]string{"EDITOR": "vim", "VISUAL": ""},
			expected: "vim",
		},
		{
			name:     "VISUAL takes precedence over EDITOR",
			envVars:  map[string]string{"EDITOR": "vim", "VISUAL": "code"},
			expected: "code",
		},
		{
			name:     "neither set, falls back to vi",
			envVars:  map[string]string{"EDITOR": "", "VISUAL": ""},
			expected: "vi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			origEditor := os.Getenv("EDITOR")
			origVisual := os.Getenv("VISUAL")
			defer func() {
				os.Setenv("EDITOR", origEditor)
				os.Setenv("VISUAL", origVisual)
			}()

			// Set test env
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			result := GetEditor()
			if result != tt.expected {
				t.Errorf("GetEditor() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateTempCommitFile(t *testing.T) {
	initial := "Initial message"
	path, err := CreateTempCommitFile(initial)
	if err != nil {
		t.Fatalf("CreateTempCommitFile() error = %v", err)
	}
	defer os.Remove(path)

	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("CreateTempCommitFile() did not create file")
	}

	// Check extension
	if ext := filepath.Ext(path); ext != ".txt" {
		t.Errorf("CreateTempCommitFile() extension = %q, want .txt", ext)
	}

	// Check content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != initial {
		t.Errorf("CreateTempCommitFile() content = %q, want %q", string(content), initial)
	}
}

func TestReadTempCommitFile(t *testing.T) {
	// Create a temp file with known content
	tmpfile, err := os.CreateTemp("", "commit*.txt")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(tmpfile.Name())

	expected := "Test commit message\n\nWith multiple lines"
	if _, err := tmpfile.WriteString(expected); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	tmpfile.Close()

	result, err := ReadTempCommitFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadTempCommitFile() error = %v", err)
	}

	if result != expected {
		t.Errorf("ReadTempCommitFile() = %q, want %q", result, expected)
	}
}

func TestReadTempCommitFile_NonExistent(t *testing.T) {
	_, err := ReadTempCommitFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadTempCommitFile() expected error for non-existent file")
	}
}
