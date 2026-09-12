package translationdemo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prismgo/framework/foundation"

	// Register the demo application configuration used by the translation provider.
	_ "prismgo-demo/config"
)

func TestTranslationDemoArchitecture(t *testing.T) {
	result := runTranslationScenario(t, "architecture")
	if result.Value != "*translation.Translator" || !hasDetail(result, "contract:translation.Translator") || !hasDetail(result, "loader:*translation.FileLoader") {
		t.Fatalf("architecture result = %#v, want Translator contract with FileLoader", result)
	}
}

func TestTranslationDemoConfiguration(t *testing.T) {
	result := runTranslationScenario(t, "config")
	if result.Value != "en" || result.Locale != "en" || !hasDetail(result, "app.fallback_locale:fr") {
		t.Fatalf("config result = %#v, want locale=en fallback=fr", result)
	}
}

func TestTranslationDemoServiceProvider(t *testing.T) {
	result := runTranslationScenario(t, "provider")
	for _, want := range []string{"translator_singleton:true", "loader_registered:true", "loader_singleton:true"} {
		if !hasDetail(result, want) {
			t.Fatalf("provider result details = %v, want %q", result.Details, want)
		}
	}
}

func TestTranslationDemoShortKeys(t *testing.T) {
	assertTranslationValue(t, "short-keys", "messages.welcome", "Welcome to PrismGo!")
}

func TestTranslationDemoNestedKeys(t *testing.T) {
	assertTranslationValue(t, "nested-keys", "messages.auth.failed", "These credentials do not match our records.")
}

func TestTranslationDemoJSONKeys(t *testing.T) {
	assertTranslationValue(t, "json-keys", "I love programming.", "I love programming.")
}

func TestTranslationDemoKeyConflicts(t *testing.T) {
	result := runTranslationScenario(t, "key-conflicts")
	if result.Value != "action" || !hasDetail(result, "has:true") || !hasDetail(result, "group.title:Action group") {
		t.Fatalf("key conflict result = %#v, want original key plus discoverable action group", result)
	}
}

func TestTranslationDemoNamespaces(t *testing.T) {
	assertTranslationValue(t, "namespaces", "acme::messages.hello", "Hello from acme!")
}

func TestTranslationDemoTranslatorInstance(t *testing.T) {
	assertTranslationValue(t, "translator", "messages.welcome", "Welcome to PrismGo!")
}

func TestTranslationDemoFacade(t *testing.T) {
	assertTranslationValue(t, "facade", "messages.welcome", "Welcome to PrismGo!")
}

func TestTranslationDemoMissingDefault(t *testing.T) {
	assertTranslationValue(t, "missing-default", "messages.missing", "messages.missing")
}

func TestTranslationDemoLocaleArgument(t *testing.T) {
	result := runTranslationScenario(t, "locale-argument")
	if result.Key != "messages.welcome" || result.Value != "Bienvenue sur PrismGo !" || result.Locale != "en" || !hasDetail(result, "requested_locale:fr") || !hasDetail(result, "current_locale_unchanged:en") {
		t.Fatalf("locale argument result = %#v, want French value with current locale unchanged at en", result)
	}
}

func TestTranslationDemoHas(t *testing.T) {
	result := runTranslationScenario(t, "has")
	if result.Value != "true" || !hasDetail(result, "missing:false") || !hasDetail(result, "group:true") {
		t.Fatalf("has result = %#v, want existing key and group true and missing key false", result)
	}
}

func TestTranslationDemoHasForLocale(t *testing.T) {
	result := runTranslationScenario(t, "has-for-locale")
	if result.Value != "true" || result.Locale != "en" || !hasDetail(result, "fr:true") || !hasDetail(result, "fr_english_only:false") {
		t.Fatalf("locale-specific existence result = %#v, want French welcome true, French english_only false, current locale en", result)
	}
}

func TestTranslationDemoGetMap(t *testing.T) {
	result := runTranslationScenario(t, "get-map")
	if result.Key != "messages" || result.Value != "Welcome to PrismGo!" || !hasDetail(result, "auth.failed:These credentials do not match our records.") {
		t.Fatalf("get-map result = %#v, want whole messages group and nested auth values", result)
	}
}

func TestTranslationDemoReplacements(t *testing.T) {
	assertTranslationValue(t, "replacements", "messages.greeting", "Hello, Taylor!")
}

func TestTranslationDemoReplacementCase(t *testing.T) {
	assertTranslationValue(t, "replacement-case", "messages.name_styles", "taylor|TAYLOR|Taylor")
}

func TestTranslationDemoStringable(t *testing.T) {
	result := runTranslationScenario(t, "stringable")
	if result.Value != "Total: $12.50 (Ada)" || !hasDetail(result, "custom_formatter:demoPrice") || !hasDetail(result, "fmt.Stringer:demoPerson") {
		t.Fatalf("stringable result = %#v, want custom demoPrice and fmt.Stringer formatting", result)
	}
}

func TestTranslationDemoPluralPipe(t *testing.T) {
	result := runTranslationScenario(t, "plural-pipe")
	if result.Value != "3 apples" || !hasDetail(result, "one:One apple") {
		t.Fatalf("plural pipe result = %#v, want one and many English variants", result)
	}
}

func TestTranslationDemoPluralIntervals(t *testing.T) {
	result := runTranslationScenario(t, "plural-intervals")
	if result.Value != "No invitations" || !hasDetail(result, "one:One invitation") || !hasDetail(result, "many:5 invitations") {
		t.Fatalf("plural intervals result = %#v, want exact zero, exact one, and open-ended many variants", result)
	}
}

func TestTranslationDemoPluralReplacements(t *testing.T) {
	assertTranslationValue(t, "plural-replacements", "messages.minutes_ago", "5 minutes ago")
}

func TestTranslationDemoPluralCount(t *testing.T) {
	result := runTranslationScenario(t, "plural-count")
	if result.Value != "There are 10 notifications" || !hasDetail(result, "custom_count:There are many notifications") {
		t.Fatalf("plural count result = %#v, want automatic count and preserved custom count replacements", result)
	}
}

func TestTranslationDemoLocale(t *testing.T) {
	result := runTranslationScenario(t, "locale")
	if result.Value != "Bienvenue sur PrismGo !" || result.Locale != "fr" || !hasDetail(result, "previous:en") || !hasDetail(result, "current:fr") || !hasDetail(result, "is_fr:true") {
		t.Fatalf("locale result = %#v, want current locale changed from en to fr", result)
	}
}

func TestTranslationDemoLocaleValidation(t *testing.T) {
	result := runTranslationScenario(t, "locale-validation")
	if result.Value != "true" || result.Locale != "en" || !hasDetail(result, "unchanged:true") || !hasDetail(result, "invalid characters in locale") {
		t.Fatalf("locale validation result = %#v, want rejected path traversal and unchanged en locale", result)
	}
}

func TestTranslationDemoFallback(t *testing.T) {
	result := runTranslationScenario(t, "fallback")
	if result.Value != "English only" || result.Locale != "es" || !hasDetail(result, "previous:fr") || !hasDetail(result, "fallback:en") {
		t.Fatalf("fallback result = %#v, want es to resolve English-only key through changed en fallback", result)
	}
}

func TestTranslationDemoFallbackValidation(t *testing.T) {
	result := runTranslationScenario(t, "fallback-validation")
	if result.Value != "true" || result.Locale != "en" || !hasDetail(result, "unchanged:true") || !hasDetail(result, "fallback:fr") || !hasDetail(result, "invalid characters in fallback locale") {
		t.Fatalf("fallback validation result = %#v, want rejected path traversal and unchanged fr fallback", result)
	}
}

func TestTranslationDemoLocaleResolver(t *testing.T) {
	result := runTranslationScenario(t, "locale-resolver")
	if result.Value != "Bienvenue sur PrismGo !" || result.Locale != "en" || !hasDetail(result, "requested:en") || !hasDetail(result, "chain:fr,en") {
		t.Fatalf("locale resolver result = %#v, want custom fr then en lookup chain for requested en", result)
	}
}

func TestTranslationDemoNamespaceOverrides(t *testing.T) {
	result := runTranslationScenario(t, "namespace-overrides")
	if result.Value != "Package fire override" || !hasDetail(result, "package_only:Package-only translation") {
		t.Fatalf("namespace override result = %#v, want hint translations to override vendor values and retain package-only values", result)
	}
}

func TestTranslationDemoAddLines(t *testing.T) {
	result := runTranslationScenario(t, "add-lines")
	if result.Value != "Hello, Taylor" || !hasDetail(result, "namespaced:Info from acme") {
		t.Fatalf("add-lines result = %#v, want runtime grouped and namespaced translations", result)
	}
}

func TestTranslationDemoAddLinesPrecedence(t *testing.T) {
	result := runTranslationScenario(t, "add-lines-precedence")
	if result.Value != "Runtime welcome" || !hasDetail(result, "file:Welcome to PrismGo!") {
		t.Fatalf("add-lines precedence result = %#v, want runtime line to override file value", result)
	}
}

func TestTranslationDemoMissingHandler(t *testing.T) {
	result := runTranslationScenario(t, "missing-handler")
	if result.Value != "Remote hello, Taylor" || !hasDetail(result, "handled_locale:en") || !hasDetail(result, "declined:messages.still_missing") {
		t.Fatalf("missing handler result = %#v, want handled value with replacements and default result after decline", result)
	}
}

func TestTranslationDemoGroupPaths(t *testing.T) {
	result := runTranslationScenario(t, "group-paths")
	if result.Value != "Loaded from an added group path" || !hasDetail(result, "override:Welcome from the later group path") {
		t.Fatalf("group paths result = %#v, want added path value and later-path override", result)
	}
}

func TestTranslationDemoJSONPaths(t *testing.T) {
	result := runTranslationScenario(t, "json-paths")
	if result.Value != "Hello from the later JSON path" || !hasDetail(result, "later_path_override:true") {
		t.Fatalf("JSON paths result = %#v, want added JSON path value and later-path override", result)
	}
}

func TestTranslationDemoCustomLoader(t *testing.T) {
	result := runTranslationScenario(t, "custom-loader")
	if result.Value != "Custom loader value" || !hasDetail(result, "loader:*translationdemo.demoLoader") || !hasDetail(result, "loads:1") {
		t.Fatalf("custom loader result = %#v, want replacement loader value and one load", result)
	}
}

func TestTranslationDemoTranslatorContract(t *testing.T) {
	result := runTranslationScenario(t, "translator-contract")
	if result.Value != "Translator contract" || !hasDetail(result, "contract:translation.Translator") || !hasDetail(result, "fallback:fr") {
		t.Fatalf("translator contract result = %#v, want interface-backed translation and fallback access", result)
	}
}

func TestTranslationDemoLoaderContract(t *testing.T) {
	result := runTranslationScenario(t, "loader-contract")
	if result.Value != "paths:1,json_paths:1" || !hasDetail(result, "namespace_registered:true") || !hasDetail(result, "duplicate_paths_ignored:true") {
		t.Fatalf("loader contract result = %#v, want path inspection, namespace inspection, and deduplication", result)
	}
}

func TestTranslationDemoSelectorContract(t *testing.T) {
	result := runTranslationScenario(t, "selector-contract")
	if result.Value != ":count results" || !hasDetail(result, "one:One result") {
		t.Fatalf("selector contract result = %#v, want plural variants selected through contract", result)
	}
}

func TestTranslationDemoReset(t *testing.T) {
	result := runTranslationScenario(t, "reset")
	if result.Value != "Welcome to PrismGo!" || !hasDetail(result, "application_scoped:true") || !hasDetail(result, "same_instance:true") {
		t.Fatalf("reset result = %#v, want facade to remain backed by the current application-scoped translator", result)
	}
}

func TestTranslationDemoIsolated(t *testing.T) {
	result := runTranslationScenario(t, "isolated")
	if result.Value != "First translator" || !hasDetail(result, "second:messages.scope") || !hasDetail(result, "independent:true") {
		t.Fatalf("isolated result = %#v, want independent translator state", result)
	}
}

func TestTranslationDemoRejectsUnknownScenario(t *testing.T) {
	basePath := t.TempDir()
	writeDemoTranslations(t, basePath)
	t.Setenv("APP_LOCALE", "en")
	t.Setenv("APP_FALLBACK_LOCALE", "fr")

	app := foundation.NewApplication(basePath)
	if err := app.Boot(); err != nil {
		t.Fatalf("boot translation demo application: %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close translation demo application: %v", err)
		}
	})

	if _, err := Run("missing", basePath); err == nil || !strings.Contains(err.Error(), `unknown scenario "missing"`) {
		t.Fatalf("Run(missing) error = %v, want unknown scenario error", err)
	}
}

func TestTranslationDemoScenariosReturnsCopy(t *testing.T) {
	first := Scenarios()
	if len(first) != 39 {
		t.Fatalf("Scenarios() returned %d entries, want 39", len(first))
	}
	first[0].Name = "changed"
	second := Scenarios()
	if second[0].Name != "architecture" {
		t.Fatalf("Scenarios()[0].Name = %q after caller mutation, want architecture", second[0].Name)
	}
}

func assertTranslationValue(t *testing.T, caseName, wantKey, wantValue string) {
	t.Helper()
	result := runTranslationScenario(t, caseName)
	if result.Case != caseName || result.Key != wantKey || result.Value != wantValue || result.Locale != "en" {
		t.Fatalf("translation scenario %q result = %#v, want key=%q value=%q locale=en", caseName, result, wantKey, wantValue)
	}
}

func runTranslationScenario(t *testing.T, caseName string) Result {
	t.Helper()
	basePath := t.TempDir()
	writeDemoTranslations(t, basePath)
	t.Setenv("APP_LOCALE", "en")
	t.Setenv("APP_FALLBACK_LOCALE", "fr")

	app := foundation.NewApplication(basePath)
	if err := app.Boot(); err != nil {
		t.Fatalf("boot translation demo application for %q: %v", caseName, err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Errorf("close translation demo application for %q: %v", caseName, err)
		}
	})

	result, err := Run(caseName, basePath)
	if err != nil {
		t.Fatalf("run translation scenario %q: %v", caseName, err)
	}
	return result
}

func writeDemoTranslations(t *testing.T, basePath string) {
	t.Helper()
	files := map[string]string{
		"lang/en/messages.json":                                `{"welcome":"Welcome to PrismGo!","english_only":"English only","greeting":"Hello, :Name!","name_styles":":name|:NAME|:Name","receipt":"Total: :price (:Name)","apples":"One apple|:count apples","invitations":"{0} No invitations|{1} One invitation|[2,*] :count invitations","minutes_ago":"{1} :value minute ago|[2,*] :value minutes ago","notifications":"{0} There are none|{1} There is one|[2,*] There are :count notifications","auth":{"failed":"These credentials do not match our records.","throttle":"Too many login attempts."}}`,
		"lang/fr/messages.json":                                `{"welcome":"Bienvenue sur PrismGo !"}`,
		"lang/en/action.json":                                  `{"title":"Action group"}`,
		"lang/en.json":                                         `{"I love programming.":"I love programming.","JSON Greeting":"Hello from the base JSON path"}`,
		"lang/vendor/acme/en/messages.json":                    `{"hello":"Hello from acme!"}`,
		"lang/vendor/hearthfire/en/messages.json":              `{"fire":"Application vendor fire"}`,
		"package-lang/hearthfire/en/messages.json":             `{"fire":"Package fire override","package_only":"Package-only translation"}`,
		"app/demo/translation/testdata/group/en/messages.json": `{"path_source":"Loaded from an added group path","welcome":"Welcome from the later group path"}`,
		"app/demo/translation/testdata/json/en.json":           `{"JSON Greeting":"Hello from the later JSON path"}`,
	}
	for name, content := range files {
		path := filepath.Join(basePath, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create parent for translation fixture %q: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write translation fixture %q: %v", name, err)
		}
	}
}

func hasDetail(result Result, want string) bool {
	return strings.Contains(strings.Join(result.Details, "\n"), want)
}
