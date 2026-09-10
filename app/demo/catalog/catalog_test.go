package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogEntries(t *testing.T) {
	entries := All()
	if err := Validate(entries); err != nil {
		t.Fatalf("validate catalog entries: %v", err)
	}
	if len(entries) != 87 {
		t.Fatalf("catalog has %d entries, want 87", len(entries))
	}
	if item, ok := Find("commands", "list"); !ok || item.Status != StatusImplemented {
		t.Fatalf("implemented demo:list entry = %#v, %v", item, ok)
	}
	if item, ok := Find("queue", "basic-sync"); !ok || item.Since != SinceInitial {
		t.Fatalf("implemented queue version = %#v, %v; want %s", item, ok, SinceInitial)
	}
	if item, ok := Find("queue", "batch-events"); !ok || item.Since != SinceInitial {
		t.Fatalf("planned demo framework version = %#v, %v; want %s", item, ok, SinceInitial)
	}
	for _, slug := range []string{
		"driver-prerequisites", "config", "payload-encoding", "sync-connection",
		"failed-store", "batch-store", "restart-store",
		"debounce-options", "job-control", "worker-command", "expiration",
		"failed-command-paths", "failed-retry", "failed-event", "batch-events",
	} {
		if item, ok := Find("queue", slug); !ok || item.Status != StatusImplemented {
			t.Fatalf("queue configuration entry %q = %#v, %v; want implemented", slug, item, ok)
		}
	}
	if got := Filter("redis", LevelIntegration, StatusPlanned); len(got) != 1 {
		t.Fatalf("redis integration filter returned %d entries, want 1", len(got))
	}
	if got := Filter("queue", "", StatusImplemented); len(got) != 38 {
		t.Fatalf("implemented queue entries = %d, want 38", len(got))
	}
	if got := Filter("queue", "", StatusPlanned); len(got) != 21 {
		t.Fatalf("planned queue entries = %d, want 21", len(got))
	}
}

func TestCatalogFeatureSummaries(t *testing.T) {
	summaries := Summaries()
	if len(summaries) != 29 {
		t.Fatalf("Summaries() returned %d features, want 29", len(summaries))
	}

	queue, ok := SummaryFor("queue")
	if !ok {
		t.Fatal("SummaryFor(queue) found = false")
	}
	if queue.Implemented != 38 || queue.Planned != 21 || queue.Manual != 0 || queue.Total != 59 || queue.Remaining != 21 {
		t.Fatalf("queue summary = %#v, want implemented=38 planned=21 manual=0 total=59 remaining=21", queue)
	}
	if queue.Status != FeatureStatusInProgress || queue.Since != SinceInitial {
		t.Fatalf("queue status/since = %q/%q, want %q/%q", queue.Status, queue.Since, FeatureStatusInProgress, SinceInitial)
	}

	commands, ok := SummaryFor("commands")
	if !ok || commands.Status != FeatureStatusImplemented {
		t.Fatalf("commands summary = %#v, %v; want implemented", commands, ok)
	}
	installation, ok := SummaryFor("installation")
	if !ok || installation.Status != FeatureStatusManual || installation.Remaining != 0 {
		t.Fatalf("installation summary = %#v, %v; want manual with no remaining planned entries", installation, ok)
	}
}

func TestFeatureStatus(t *testing.T) {
	tests := []struct {
		name    string
		summary Summary
		want    FeatureStatus
	}{
		{name: "implemented", summary: Summary{Implemented: 2, Total: 2}, want: FeatureStatusImplemented},
		{name: "in progress", summary: Summary{Implemented: 1, Planned: 1, Total: 2}, want: FeatureStatusInProgress},
		{name: "planned", summary: Summary{Planned: 2, Total: 2}, want: FeatureStatusPlanned},
		{name: "manual", summary: Summary{Manual: 2, Total: 2}, want: FeatureStatusManual},
		{name: "empty", summary: Summary{}, want: FeatureStatusPlanned},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := featureStatus(test.summary); got != test.want {
				t.Fatalf("featureStatus(%#v) = %q, want %q", test.summary, got, test.want)
			}
		})
	}
}

func TestCatalogValidationRejectsInvalidEntries(t *testing.T) {
	valid := All()[0]
	tests := []struct {
		name  string
		items []Entry
	}{
		{name: "empty", items: []Entry{{}}},
		{name: "since", items: []Entry{func() Entry { item := valid; item.Since = ""; return item }()}},
		{name: "feature metadata", items: []Entry{func() Entry { item := valid; item.Feature = "missing"; return item }()}},
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
	if err := ValidateDocuments(docsRoot, All()); err != nil {
		t.Fatalf("validate catalog documents: %v", err)
	}
}

func TestFindDocsRootAndMissingHeading(t *testing.T) {
	root := t.TempDir()
	docsRoot := filepath.Join(root, "docs")
	for _, locale := range []string{"zh_CN", "en"} {
		if err := os.MkdirAll(filepath.Join(docsRoot, locale), 0o755); err != nil {
			t.Fatalf("create %s docs directory: %v", locale, err)
		}
		if err := os.WriteFile(filepath.Join(docsRoot, locale, "sample.md"), []byte("# Sample\n\n## Present\n"), 0o644); err != nil {
			t.Fatalf("write %s sample document: %v", locale, err)
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
