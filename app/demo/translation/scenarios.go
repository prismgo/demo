// Package translationdemo contains runnable examples for the public translation interface.
package translationdemo

// Scenario describes one documented translation behavior exposed by demo:translation.
type Scenario struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	Status  string `json:"status"`
}

var scenarios = []Scenario{
	{Name: "architecture", Section: "Overview", Status: "implemented"},
	{Name: "config", Section: "Environment Variables", Status: "implemented"},
	{Name: "provider", Section: "Registering the Service Provider", Status: "implemented"},
	{Name: "short-keys", Section: "Short Keys", Status: "implemented"},
	{Name: "nested-keys", Section: "Short Keys", Status: "implemented"},
	{Name: "json-keys", Section: "JSON Keys", Status: "implemented"},
	{Name: "key-conflicts", Section: "JSON Keys", Status: "implemented"},
	{Name: "namespaces", Section: "Namespaces", Status: "implemented"},
	{Name: "translator", Section: "Using a Translator Instance", Status: "implemented"},
	{Name: "facade", Section: "Using Facade Helpers", Status: "implemented"},
	{Name: "missing-default", Section: "Using Facade Helpers", Status: "implemented"},
	{Name: "locale-argument", Section: "Specifying a Locale", Status: "implemented"},
	{Name: "has", Section: "Checking Whether a Key Exists", Status: "implemented"},
	{Name: "has-for-locale", Section: "Checking Whether a Key Exists", Status: "implemented"},
	{Name: "get-map", Section: "Retrieving a Whole Group", Status: "implemented"},
	{Name: "replacements", Section: "Replacing Parameters", Status: "implemented"},
	{Name: "replacement-case", Section: "Placeholder Case Rules", Status: "implemented"},
	{Name: "stringable", Section: "Custom Formatting With Stringable", Status: "implemented"},
	{Name: "plural-pipe", Section: "Basic Pipe Syntax", Status: "implemented"},
	{Name: "plural-intervals", Section: "Explicit Intervals", Status: "implemented"},
	{Name: "plural-replacements", Section: "Placeholders in Pluralized Strings", Status: "implemented"},
	{Name: "plural-count", Section: "Placeholders in Pluralized Strings", Status: "implemented"},
	{Name: "locale", Section: "Current Locale", Status: "implemented"},
	{Name: "locale-validation", Section: "Current Locale", Status: "implemented"},
	{Name: "fallback", Section: "Fallback Locale", Status: "implemented"},
	{Name: "fallback-validation", Section: "Fallback Locale", Status: "implemented"},
	{Name: "locale-resolver", Section: "Custom Locale Resolution", Status: "implemented"},
	{Name: "namespace-overrides", Section: "Overriding Package Language Files", Status: "implemented"},
	{Name: "add-lines", Section: "Adding Lines at Runtime", Status: "implemented"},
	{Name: "add-lines-precedence", Section: "Adding Lines at Runtime", Status: "implemented"},
	{Name: "missing-handler", Section: "Handling Missing Keys", Status: "implemented"},
	{Name: "group-paths", Section: "Group Paths", Status: "implemented"},
	{Name: "json-paths", Section: "JSON Paths", Status: "implemented"},
	{Name: "custom-loader", Section: "Custom Loaders", Status: "implemented"},
	{Name: "translator-contract", Section: "Translator", Status: "implemented"},
	{Name: "loader-contract", Section: "Loader", Status: "implemented"},
	{Name: "selector-contract", Section: "Selector", Status: "implemented"},
	{Name: "reset", Section: "Testing", Status: "implemented"},
	{Name: "isolated", Section: "Testing", Status: "implemented"},
}

// Scenarios returns a copy of the implemented translation scenarios in presentation order.
func Scenarios() []Scenario {
	return append([]Scenario(nil), scenarios...)
}
