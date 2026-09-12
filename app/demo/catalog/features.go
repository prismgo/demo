package catalog

import "sort"

// SinceInitial marks capabilities present in the first tagged framework release.
const SinceInitial = "v0.1.0"

// FeatureStatus summarizes implementation progress across a feature's entries.
type FeatureStatus string

const (
	// FeatureStatusImplemented means every catalog entry is implemented.
	FeatureStatusImplemented FeatureStatus = "implemented"
	// FeatureStatusInProgress means at least one, but not every, entry is implemented.
	FeatureStatusInProgress FeatureStatus = "in_progress"
	// FeatureStatusPlanned means the feature still has only planned implementation entries.
	FeatureStatusPlanned FeatureStatus = "planned"
	// FeatureStatusManual means every entry is verified manually or externally.
	FeatureStatusManual FeatureStatus = "manual"
)

// Feature describes one top-level framework documentation module.
type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Since       string `json:"since"`
}

// Summary reports implementation progress for one feature.
type Summary struct {
	Feature     string        `json:"feature"`
	Description string        `json:"description"`
	Since       string        `json:"since"`
	Implemented int           `json:"implemented"`
	Planned     int           `json:"planned"`
	Manual      int           `json:"manual"`
	Total       int           `json:"total"`
	Remaining   int           `json:"remaining"`
	Status      FeatureStatus `json:"status"`
}

var features = []Feature{
	{Name: "cache", Description: "Cache stores, tags, and locks", Since: SinceInitial},
	{Name: "commands", Description: "Application commands and options", Since: SinceInitial},
	{Name: "config", Description: "Environment and application configuration", Since: SinceInitial},
	{Name: "console", Description: "Artisan-style commands and terminal IO", Since: SinceInitial},
	{Name: "container", Description: "Dependency binding and resolution", Since: SinceInitial},
	{Name: "cookie", Description: "Cookie values and queued writes", Since: SinceInitial},
	{Name: "database", Description: "Connections, models, migrations, and indexes", Since: SinceInitial},
	{Name: "encryption", Description: "Application data encryption", Since: SinceInitial},
	{Name: "event", Description: "Synchronous, asynchronous, and queued events", Since: SinceInitial},
	{Name: "exception", Description: "Exception reporting and rendering", Since: SinceInitial},
	{Name: "facade", Description: "Framework facade access", Since: SinceInitial},
	{Name: "filesystem", Description: "Local, public, and cloud filesystems", Since: SinceInitial},
	{Name: "horizon", Description: "Queue monitoring and worker management", Since: SinceInitial},
	{Name: "http-server", Description: "HTTP server startup and lifecycle", Since: SinceInitial},
	{Name: "installation", Description: "Application installation workflow", Since: SinceInitial},
	{Name: "lens", Description: "Lens development workflow", Since: SinceInitial},
	{Name: "lifecycle", Description: "Application bootstrap and shutdown", Since: SinceInitial},
	{Name: "logger", Description: "Multi-channel application logging", Since: SinceInitial},
	{Name: "queue", Description: "Queues, jobs, and workers", Since: SinceInitial},
	{Name: "ratelimit", Description: "Request and action rate limiting", Since: SinceInitial},
	{Name: "redis", Description: "Redis connections and operations", Since: SinceInitial},
	{Name: "route", Description: "HTTP route registration", Since: SinceInitial},
	{Name: "schema", Description: "Database schema and migration builder", Since: SinceInitial},
	{Name: "service-provider", Description: "Service provider registration and lifecycle", Since: SinceInitial},
	{Name: "session", Description: "Server-side session storage", Since: SinceInitial},
	{Name: "starter", Description: "Generated application starter", Since: SinceInitial},
	{Name: "support", Description: "General framework helpers", Since: SinceInitial},
	{Name: "timer", Description: "Scheduled task definitions", Since: SinceInitial},
	{Name: "translation", Description: "Application translation and pluralization", Since: SinceInitial},
}

// Features returns top-level feature metadata in name order.
func Features() []Feature {
	result := append([]Feature(nil), features...)
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// LookupFeature resolves top-level metadata by its stable feature name.
func LookupFeature(name string) (Feature, bool) {
	for _, feature := range features {
		if feature.Name == name {
			return feature, true
		}
	}
	return Feature{}, false
}

// Summaries returns implementation summaries in feature name order.
func Summaries() []Summary {
	byFeature := make(map[string]*Summary, len(features))
	for _, feature := range Features() {
		byFeature[feature.Name] = &Summary{
			Feature: feature.Name, Description: feature.Description, Since: feature.Since,
		}
	}
	for _, item := range All() {
		summary, ok := byFeature[item.Feature]
		if !ok {
			continue
		}
		summary.Total++
		switch item.Status {
		case StatusImplemented:
			summary.Implemented++
		case StatusPlanned:
			summary.Planned++
		case StatusManual:
			summary.Manual++
		}
	}

	result := make([]Summary, 0, len(byFeature))
	for _, feature := range Features() {
		summary := byFeature[feature.Name]
		summary.Remaining = summary.Planned
		summary.Status = featureStatus(*summary)
		result = append(result, *summary)
	}
	return result
}

// SummaryFor returns the implementation summary for one feature.
func SummaryFor(name string) (Summary, bool) {
	for _, summary := range Summaries() {
		if summary.Feature == name {
			return summary, true
		}
	}
	return Summary{}, false
}

func featureStatus(summary Summary) FeatureStatus {
	if summary.Total > 0 && summary.Implemented == summary.Total {
		return FeatureStatusImplemented
	}
	if summary.Implemented > 0 {
		return FeatureStatusInProgress
	}
	if summary.Total > 0 && summary.Manual == summary.Total {
		return FeatureStatusManual
	}
	return FeatureStatusPlanned
}
