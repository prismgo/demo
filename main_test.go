package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalApplicationBasePath(t *testing.T) {
	workspace := t.TempDir()
	demoDirectory := filepath.Join(workspace, "demo")
	if err := os.Mkdir(demoDirectory, 0o755); err != nil {
		t.Fatalf("mkdir demo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(demoDirectory, "go.mod"), []byte("module example.test/demo\n"), 0o644); err != nil {
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
	if got := localApplicationBasePath(); got != demoDirectory {
		t.Fatalf("localApplicationBasePath() = %q, want %q", got, demoDirectory)
	}

	if err := os.Chdir(demoDirectory); err != nil {
		t.Fatalf("chdir demo: %v", err)
	}
	if got := localApplicationBasePath(); got != "" {
		t.Fatalf("localApplicationBasePath() from demo = %q, want empty auto-detect path", got)
	}
}
