package providerdemo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	providerpkg "github.com/prismgo/framework/provider"
)

// publishesScenario verifies vendor resource publication registration and copy.
func publishesScenario(base string) (string, error) {
	providerpkg.PublishClear()
	defer providerpkg.PublishClear()

	source, err := writePublishFixture(base, "config/demo.php", "<?php return [];\n")
	if err != nil {
		return "", err
	}
	target := filepath.Join(base, "published", "config", "demo.php")
	if err := providerpkg.Publishes("demo.acme", map[string]string{source: target}, "config"); err != nil {
		return "", fmt.Errorf("register demo.acme publish: %w", err)
	}
	entries := providerpkg.PublishEntries("demo.acme", nil)
	if _, _, err := providerpkg.PublishCopy("demo.acme", nil, false, false); err != nil {
		return "", fmt.Errorf("copy demo.acme publish: %w", err)
	}
	copied := false
	if _, err := os.Stat(target); err == nil {
		copied = true
	}
	tags := ""
	if len(entries) > 0 {
		tags = strings.Join(entries[0].Tags, ",")
	}
	return fmt.Sprintf("entries=%d tags=%s copied=%t", len(entries), tags, copied), nil
}

// publishTagsScenario verifies provider and tag filtering of published resources.
func publishTagsScenario(base string) (string, error) {
	providerpkg.PublishClear()
	defer providerpkg.PublishClear()

	configSource, err := writePublishFixture(base, "config/service.php", "<?php return [];\n")
	if err != nil {
		return "", err
	}
	if err := providerpkg.Publishes("demo.config", map[string]string{configSource: filepath.Join(base, "published", "service.php")}, "config"); err != nil {
		return "", fmt.Errorf("register demo.config publish: %w", err)
	}
	langSource, err := writePublishFixture(base, "lang/en/messages.php", "<?php return [];\n")
	if err != nil {
		return "", err
	}
	if err := providerpkg.Publishes("demo.lang", map[string]string{langSource: filepath.Join(base, "published", "messages.php")}, "lang"); err != nil {
		return "", fmt.Errorf("register demo.lang publish: %w", err)
	}

	providers := providerpkg.PublishProviders()
	tags := providerpkg.PublishTags()
	filtered := providerpkg.PublishEntries("", []string{"config"})
	return fmt.Sprintf("providers=%s tags=%s filtered=%d",
		strings.Join(providers, ","), strings.Join(tags, ","), len(filtered)), nil
}

// publishEnvironmentScenario verifies publication is disabled in production.
func publishEnvironmentScenario(base string) (string, error) {
	providerpkg.PublishClear()
	defer providerpkg.PublishClear()

	previous, had := os.LookupEnv("APP_ENV")
	if err := os.Setenv("APP_ENV", "production"); err != nil {
		return "", fmt.Errorf("set production environment: %w", err)
	}
	defer func() {
		if had {
			_ = os.Setenv("APP_ENV", previous)
			return
		}
		_ = os.Unsetenv("APP_ENV")
	}()

	source, err := writePublishFixture(base, "config/prod.php", "<?php return [];\n")
	if err != nil {
		return "", err
	}
	if err := providerpkg.Publishes("demo.prod", map[string]string{source: filepath.Join(base, "published", "prod.php")}, "config"); err != nil {
		return "", fmt.Errorf("register demo.prod publish: %w", err)
	}
	entries := providerpkg.PublishEntries("", nil)
	available := providerpkg.PublishIsAvailable()
	_, _, copyErr := providerpkg.PublishCopy("demo.prod", nil, false, false)
	return fmt.Sprintf("available=%t entries=%d copy-error=%t", available, len(entries), copyErr != nil), nil
}

// writePublishFixture writes a source file under the scenario base directory.
func writePublishFixture(base, relative, content string) (string, error) {
	path := filepath.Join(base, "publish-src", relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create publish fixture directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write publish fixture %s: %w", relative, err)
	}
	return path, nil
}
