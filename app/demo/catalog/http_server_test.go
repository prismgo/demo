package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTTPServerDocumentationCoverage(t *testing.T) {
	entries := Filter("http-server", "", "")
	if len(entries) != 29 {
		t.Fatalf("HTTP server catalog entries = %d, want 29", len(entries))
	}
	for _, caseName := range []string{
		"bootstrap", "routes", "serve", "port-flag", "port-flag-scope", "host", "port",
		"timeout-fallback", "read-timeout", "read-header-timeout", "write-timeout", "idle-timeout",
		"shutdown-timeout", "max-header-bytes", "max-multipart-memory", "duration-string", "duration-seconds",
		"trusted-proxies", "client-ip-headers", "untrusted-client-ip", "access-log", "exception-handler",
		"debug-exception", "pid-file", "stop", "kill", "reload", "restart", "missing-pid",
	} {
		if _, found := Find("http-server", caseName); !found {
			t.Errorf("HTTP server catalog case %q found = false, want true", caseName)
		}
	}
	for _, item := range entries {
		if item.Status != StatusImplemented {
			t.Errorf("HTTP server entry %q status = %q, want %q", item.Case, item.Status, StatusImplemented)
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
		{path: filepath.Join(docsRoot, "zh_CN", "http-server.md"), heading: func(item Entry) string { return item.HeadingZH }},
		{path: filepath.Join(docsRoot, "en", "http-server.md"), heading: func(item Entry) string { return item.HeadingEN }},
	} {
		content, err := os.ReadFile(locale.path)
		if err != nil {
			t.Fatalf("read HTTP server documentation %s: %v", locale.path, err)
		}
		configRows := 0
		for _, rawLine := range strings.Split(string(content), "\n") {
			line := strings.TrimSpace(rawLine)
			if strings.HasPrefix(line, "## ") {
				heading := strings.TrimSpace(strings.TrimPrefix(line, "## "))
				if !hasHTTPServerHeading(entries, locale.heading, heading) {
					t.Errorf("HTTP server document %s heading %q has no catalog entry; want heading mapped", locale.path, heading)
				}
			}
			if !strings.HasPrefix(line, "| `app.server.") {
				continue
			}
			columns := strings.Split(line, "|")
			if len(columns) < 4 {
				t.Fatalf("HTTP server document %s row %q has %d columns, want at least 4", locale.path, line, len(columns))
			}
			section := strings.Trim(strings.TrimSpace(columns[1]), "`") + " and " + strings.Trim(strings.TrimSpace(columns[2]), "`")
			if !hasHTTPServerSection(entries, section) {
				t.Errorf("HTTP server document %s row %q has no catalog entry; want section %q", locale.path, line, section)
			}
			configRows++
		}
		if configRows != 14 {
			t.Errorf("HTTP server document %s has %d configuration rows, want 14", locale.path, configRows)
		}
		for _, command := range []string{"go run . serve", "--port=8000", "--stop", "--kill", "--reload", "--restart"} {
			if !strings.Contains(string(content), command) {
				t.Errorf("HTTP server document %s is missing command %q; want documented command", locale.path, command)
			}
		}
	}
}

func hasHTTPServerHeading(entries []Entry, headingFor func(Entry) string, heading string) bool {
	for _, item := range entries {
		if headingFor(item) == heading {
			return true
		}
	}
	return false
}

func hasHTTPServerSection(entries []Entry, section string) bool {
	for _, item := range entries {
		if item.Section == section {
			return true
		}
	}
	return false
}
