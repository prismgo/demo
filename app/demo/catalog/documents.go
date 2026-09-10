package catalog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindDocsRoot walks upward from start and returns a sibling docs checkout when
// available. A standalone demo checkout legitimately returns ok=false.
func FindDocsRoot(start string) (root string, ok bool) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		candidate := filepath.Join(current, "docs")
		if isDirectory(filepath.Join(candidate, "zh_CN")) && isDirectory(filepath.Join(candidate, "en")) {
			return candidate, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

// ValidateDocuments verifies localized file paths and headings.
func ValidateDocuments(docsRoot string, entries []Entry) error {
	for _, item := range entries {
		for _, document := range []struct {
			path    string
			heading string
		}{
			{path: item.DocumentZH, heading: item.HeadingZH},
			{path: item.DocumentEN, heading: item.HeadingEN},
		} {
			path := filepath.Join(docsRoot, filepath.FromSlash(document.path))
			found, err := documentHasHeading(path, document.heading)
			if err != nil {
				return fmt.Errorf("catalog document %s: %w", document.path, err)
			}
			if !found {
				return fmt.Errorf("catalog document %s does not contain heading %q", document.path, document.heading)
			}
		}
	}
	return nil
}

func documentHasHeading(path, wanted string) (found bool, err error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close catalog document: %w", closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if heading == wanted {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
