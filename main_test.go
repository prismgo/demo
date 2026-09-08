package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalApplicationBasePath(t *testing.T) {
	workspace := t.TempDir()
	devDirectory := filepath.Join(workspace, "dev")
	if err := os.Mkdir(devDirectory, 0o755); err != nil {
		t.Fatalf("mkdir dev: %v", err)
	}
	if err := os.WriteFile(filepath.Join(devDirectory, "go.mod"), []byte("module example.test/dev\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	previousDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir workspace: %v", err)
	}
	if got := localApplicationBasePath(); got != devDirectory {
		t.Fatalf("localApplicationBasePath() = %q, want %q", got, devDirectory)
	}

	if err := os.Chdir(devDirectory); err != nil {
		t.Fatalf("chdir dev: %v", err)
	}
	if got := localApplicationBasePath(); got != "" {
		t.Fatalf("localApplicationBasePath() from dev = %q, want empty auto-detect path", got)
	}
}
