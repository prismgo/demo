package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogEntries(t *testing.T) {
	if err := Validate(Entries); err != nil {
		t.Fatal(err)
	}
	if len(Entries) != 29 {
		t.Fatalf("catalog has %d topics, want 29", len(Entries))
	}
	if item, ok := Find("commands", "list"); !ok || item.Status != StatusImplemented {
		t.Fatalf("implemented demo:list entry = %#v, %v", item, ok)
	}
	if got := Filter("redis", LevelIntegration, StatusPlanned); len(got) != 1 {
		t.Fatalf("redis integration filter returned %d entries, want 1", len(got))
	}
}

func TestCatalogValidationRejectsInvalidEntries(t *testing.T) {
	valid := Entries[0]
	tests := []struct {
		name  string
		items []Entry
	}{
		{name: "empty", items: []Entry{{}}},
		{name: "level", items: []Entry{func() Entry { item := valid; item.Level = "unknown"; return item }()}},
		{name: "status", items: []Entry{func() Entry { item := valid; item.Status = "unknown"; return item }()}},
		{name: "duplicate", items: []Entry{valid, valid}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Validate(test.items); err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}

func TestCatalogDocuments(t *testing.T) {
	docsRoot, ok := FindDocsRoot(".")
	if !ok {
		t.Skip("sibling docs checkout is not available")
	}
	if err := ValidateDocuments(docsRoot, Entries); err != nil {
		t.Fatal(err)
	}
}

func TestFindDocsRootAndMissingHeading(t *testing.T) {
	root := t.TempDir()
	docsRoot := filepath.Join(root, "docs")
	for _, locale := range []string{"zh_CN", "en"} {
		if err := os.MkdirAll(filepath.Join(docsRoot, locale), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(docsRoot, locale, "sample.md"), []byte("# Sample\n\n## Present\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	found, ok := FindDocsRoot(filepath.Join(root, "nested"))
	if !ok || found != docsRoot {
		t.Fatalf("FindDocsRoot() = %q, %v; want %q, true", found, ok, docsRoot)
	}
	item := Entry{DocumentZH: "zh_CN/sample.md", DocumentEN: "en/sample.md", HeadingZH: "Missing", HeadingEN: "Present"}
	if err := ValidateDocuments(docsRoot, []Entry{item}); err == nil || !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("missing heading error = %v", err)
	}
}
