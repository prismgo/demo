// Package catalog maps public documentation topics to runnable demos and tests.
package catalog

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Level describes how a documentation example is verified.
type Level string

const (
	// LevelCompile verifies that an example compiles.
	LevelCompile Level = "compile"
	// LevelHermetic verifies an example without external services.
	LevelHermetic Level = "hermetic"
	// LevelScenario verifies an application-level scenario.
	LevelScenario Level = "scenario"
	// LevelIntegration verifies an example against external services.
	LevelIntegration Level = "integration"
)

// Status distinguishes delivered coverage from planned and manual checks.
type Status string

const (
	// StatusImplemented marks delivered coverage.
	StatusImplemented Status = "implemented"
	// StatusPlanned marks coverage that has not been implemented.
	StatusPlanned Status = "planned"
	// StatusManual marks coverage verified outside the automated catalog.
	StatusManual Status = "manual"
)

// Entry maps one documentation topic to its demo and verification target.
// HeadingZH and HeadingEN hold the real localized Markdown headings while
// Section remains a stable, language-neutral catalog label.
type Entry struct {
	Feature      string   `json:"feature"`
	Since        string   `json:"since"`
	DocumentZH   string   `json:"document_zh"`
	DocumentEN   string   `json:"document_en"`
	Section      string   `json:"section"`
	HeadingZH    string   `json:"heading_zh"`
	HeadingEN    string   `json:"heading_en"`
	Example      string   `json:"example"`
	Case         string   `json:"case"`
	Test         string   `json:"test"`
	Level        Level    `json:"level"`
	Status       Status   `json:"status"`
	Requirements []string `json:"requirements,omitempty"`
	Notes        string   `json:"notes,omitempty"`
}

// Entries is the initial documentation coverage baseline. Planned entries are
// intentionally visible: demo:list summarizes them and demo:show remains the gap report.
var Entries = slices.Concat(
	cacheEntries(),
	commandsEntries(),
	configEntries(),
	consoleEntries(),
	containerEntries(),
	cookieEntries(),
	databaseEntries(),
	encryptionEntries(),
	eventEntries(),
	exceptionEntries(),
	facadeEntries(),
	filesystemEntries(),
	horizonEntries(),
	httpServerEntries(),
	installationEntries(),
	lensEntries(),
	lifecycleEntries(),
	loggerEntries(),
	queueEntries(),
	ratelimitEntries(),
	redisEntries(),
	routeEntries(),
	schemaEntries(),
	serviceProviderEntries(),
	sessionEntries(),
	starterEntries(),
	supportEntries(),
	timerEntries(),
	translationEntries(),
)

func baselineEntry(feature, section, headingZH, headingEN, example, caseName, test string, level Level, status Status, requirements ...string) Entry {
	return versionedEntry(SinceInitial, feature, section, headingZH, headingEN, example, caseName, test, level, status, requirements...)
}

// versionedEntry keeps release provenance explicit for capabilities added after
// the initial catalog baseline.
func versionedEntry(since, feature, section, headingZH, headingEN, example, caseName, test string, level Level, status Status, requirements ...string) Entry {
	return Entry{
		Feature: feature, Since: since, DocumentZH: "zh_CN/" + feature + ".md", DocumentEN: "en/" + feature + ".md",
		Section: section, HeadingZH: headingZH, HeadingEN: headingEN, Example: example, Case: caseName,
		Test: test, Level: level, Status: status, Requirements: requirements,
	}
}

func manualEntry(feature, section, headingZH, headingEN, example, test string, level Level, notes string) Entry {
	item := baselineEntry(feature, section, headingZH, headingEN, example, "manual", test, level, StatusManual)
	item.Notes = notes
	return item
}

// All returns a copy sorted by feature and case.
func All() []Entry {
	result := make([]Entry, len(Entries))
	for i, item := range Entries {
		result[i] = cloneEntry(item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Feature == result[j].Feature {
			return result[i].Case < result[j].Case
		}
		return result[i].Feature < result[j].Feature
	})
	return result
}

// Filter returns catalog entries matching non-empty filters.
func Filter(feature string, level Level, status Status) []Entry {
	feature = strings.TrimSpace(feature)
	result := make([]Entry, 0)
	for _, item := range All() {
		if feature != "" && item.Feature != feature {
			continue
		}
		if level != "" && item.Level != level {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		result = append(result, item)
	}
	return result
}

// Find resolves one feature/case pair.
func Find(feature, caseName string) (Entry, bool) {
	for _, item := range Entries {
		if item.Feature == feature && item.Case == caseName {
			return cloneEntry(item), true
		}
	}
	return Entry{}, false
}

func cloneEntry(item Entry) Entry {
	item.Requirements = append([]string(nil), item.Requirements...)
	return item
}

// Validate checks catalog completeness and uniqueness without accessing docs.
func Validate(entries []Entry) error {
	seen := make(map[string]struct{}, len(entries))
	for index, item := range entries {
		if strings.TrimSpace(item.Feature) == "" || strings.TrimSpace(item.Section) == "" ||
			strings.TrimSpace(item.Example) == "" || strings.TrimSpace(item.Test) == "" || strings.TrimSpace(item.Case) == "" ||
			strings.TrimSpace(item.Since) == "" {
			return fmt.Errorf("catalog entry %d has an empty required field", index)
		}
		if _, ok := LookupFeature(item.Feature); !ok {
			return fmt.Errorf("catalog entry %s/%s has unknown feature metadata", item.Feature, item.Case)
		}
		if strings.TrimSpace(item.DocumentZH) == "" || strings.TrimSpace(item.DocumentEN) == "" ||
			strings.TrimSpace(item.HeadingZH) == "" || strings.TrimSpace(item.HeadingEN) == "" {
			return fmt.Errorf("catalog entry %s/%s has incomplete documentation mapping", item.Feature, item.Case)
		}
		if !validLevel(item.Level) {
			return fmt.Errorf("catalog entry %s/%s has unknown level %q", item.Feature, item.Case, item.Level)
		}
		if !validStatus(item.Status) {
			return fmt.Errorf("catalog entry %s/%s has unknown status %q", item.Feature, item.Case, item.Status)
		}
		key := item.Feature + "\x00" + item.Case
		if _, exists := seen[key]; exists {
			return fmt.Errorf("catalog contains duplicate feature/case %s/%s", item.Feature, item.Case)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validLevel(level Level) bool {
	return level == LevelCompile || level == LevelHermetic || level == LevelScenario || level == LevelIntegration
}

func validStatus(status Status) bool {
	return status == StatusImplemented || status == StatusPlanned || status == StatusManual
}
