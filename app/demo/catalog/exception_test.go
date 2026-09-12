package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExceptionDocumentationCoverage(t *testing.T) {
	entries := Filter("exception", "", "")
	if len(entries) != 77 {
		t.Fatalf("exception catalog entries = %d, want 77", len(entries))
	}
	implemented := 0
	for _, item := range entries {
		if item.Status == StatusImplemented {
			implemented++
		}
	}
	if implemented != 77 {
		t.Fatalf("implemented exception entries = %d, want 77", implemented)
	}
	for index, item := range exceptionEntries() {
		want := StatusImplemented
		if item.Status != want {
			t.Errorf("exception entry %q status = %q, want %q (index %d)", item.Case, item.Status, want, index)
		}
	}

	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}
	for _, locale := range []struct {
		path    string
		heading func(Entry) string
	}{
		{path: filepath.Join(docsRoot, "zh_CN", "exception.md"), heading: func(item Entry) string { return item.HeadingZH }},
		{path: filepath.Join(docsRoot, "en", "exception.md"), heading: func(item Entry) string { return item.HeadingEN }},
	} {
		content, err := os.ReadFile(locale.path)
		if err != nil {
			t.Fatalf("read exception documentation %s: %v", locale.path, err)
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			depth := len(line) - len(strings.TrimLeft(line, "#"))
			if depth < 2 || depth > 4 || len(line) <= depth || line[depth] != ' ' {
				continue
			}
			hasChild := false
			for _, next := range lines[i+1:] {
				next = strings.TrimSpace(next)
				nextDepth := len(next) - len(strings.TrimLeft(next, "#"))
				if nextDepth < 2 || nextDepth > 4 || len(next) <= nextDepth || next[nextDepth] != ' ' {
					continue
				}
				hasChild = nextDepth > depth
				break
			}
			if hasChild {
				continue
			}
			heading := strings.TrimSpace(line[depth:])
			found := false
			for _, item := range entries {
				if locale.heading(item) == heading {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("exception document %s leaf heading %q has no catalog entry; actual entries=%d, want matching heading", locale.path, heading, len(entries))
			}
		}
	}

	for _, option := range []string{
		"WithDontReport", "WithLevel", "WithReporter", "WithRecovery", "WithLogging", "WithClientErrorLogging",
		"WithPanicStack", "WithDebug", "WithDebugResolver", "WithContext", "WithRenderer", "WithResponseRenderer",
	} {
		found := false
		for _, item := range entries {
			if strings.HasPrefix(item.Section, option+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("exception option %q catalog coverage = false, want true", option)
		}
	}
}
