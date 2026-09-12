package translationdemo

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	transcontract "github.com/prismgo/framework/contracts/translation"
	frameworktranslation "github.com/prismgo/framework/translation"
)

var (
	_ transcontract.Translator = (*frameworktranslation.Translator)(nil)
	_ transcontract.Loader     = (*frameworktranslation.FileLoader)(nil)
	_ transcontract.Loader     = (*demoLoader)(nil)
	_ transcontract.Selector   = (*frameworktranslation.MessageSelector)(nil)
)

// Result is the public observation returned by a translation demo scenario.
type Result struct {
	Case    string   `json:"case"`
	Key     string   `json:"key"`
	Value   string   `json:"value"`
	Locale  string   `json:"locale"`
	Details []string `json:"details,omitempty"`
}

// Run executes one translation demo scenario against the booted application translator.
func Run(caseName, basePath string) (Result, error) {
	translator := frameworktranslation.Resolve()
	if translator == nil {
		return Result{}, fmt.Errorf("translation demo: translator is nil")
	}

	switch caseName {
	case "architecture":
		return architectureResult(translator), nil
	case "config":
		return configResult(translator), nil
	case "provider":
		return providerResult(translator), nil
	case "short-keys":
		return translatedResult(caseName, "messages.welcome", translator.Get("messages.welcome", nil), translator.Locale()), nil
	case "nested-keys":
		return nestedKeysResult(translator)
	case "json-keys":
		const key = "I love programming."
		return translatedResult(caseName, key, translator.Get(key, nil), translator.Locale()), nil
	case "key-conflicts":
		return keyConflictResult(translator), nil
	case "namespaces":
		return namespaceResult(translator, basePath), nil
	case "translator":
		return translatedResult(caseName, "messages.welcome", translator.Get("messages.welcome", nil), translator.Locale()), nil
	case "facade":
		return translatedResult(caseName, "messages.welcome", frameworktranslation.Get("messages.welcome", nil), frameworktranslation.Locale()), nil
	case "missing-default":
		const key = "messages.missing"
		return translatedResult(caseName, key, translator.Get(key, nil), translator.Locale()), nil
	case "locale-argument":
		return localeArgumentResult(translator), nil
	case "has":
		return hasResult(translator), nil
	case "has-for-locale":
		return hasForLocaleResult(translator), nil
	case "get-map":
		return getMapResult(translator)
	case "replacements":
		return translatedResult(caseName, "messages.greeting", translator.Get("messages.greeting", map[string]any{"name": "Taylor"}), translator.Locale()), nil
	case "replacement-case":
		return translatedResult(caseName, "messages.name_styles", translator.Get("messages.name_styles", map[string]any{"name": "tAyLoR"}), translator.Locale()), nil
	case "stringable":
		return stringableResult(translator), nil
	case "plural-pipe":
		return pluralPipeResult(translator), nil
	case "plural-intervals":
		return pluralIntervalsResult(translator), nil
	case "plural-replacements":
		return pluralReplacementsResult(translator), nil
	case "plural-count":
		return pluralCountResult(translator), nil
	case "locale":
		return localeResult(translator)
	case "locale-validation":
		return localeValidationResult(translator), nil
	case "fallback":
		return fallbackResult(translator)
	case "fallback-validation":
		return fallbackValidationResult(translator), nil
	case "locale-resolver":
		return localeResolverResult(translator), nil
	case "namespace-overrides":
		return namespaceOverridesResult(translator, basePath), nil
	case "add-lines":
		return addLinesResult(translator), nil
	case "add-lines-precedence":
		return addLinesPrecedenceResult(translator), nil
	case "missing-handler":
		return missingHandlerResult(translator), nil
	case "group-paths":
		return groupPathsResult(translator, basePath), nil
	case "json-paths":
		return jsonPathsResult(translator, basePath), nil
	case "custom-loader":
		return customLoaderResult(translator)
	case "translator-contract":
		return translatorContractResult(), nil
	case "loader-contract":
		return loaderContractResult(basePath), nil
	case "selector-contract":
		return selectorContractResult(), nil
	case "reset":
		return resetResult(translator), nil
	case "isolated":
		return isolatedResult(), nil
	default:
		return Result{}, fmt.Errorf("translation demo: unknown scenario %q", caseName)
	}
}

func missingHandlerResult(translator transcontract.Translator) Result {
	var handledLocale string
	translator.HandleMissingKeysUsing(func(_ context.Context, key, locale string) (string, bool) {
		handledLocale = locale
		if key == "messages.remote" {
			return "Remote hello, :Name", true
		}
		return "", false
	})

	const key = "messages.remote"
	return Result{
		Case:    "missing-handler",
		Key:     key,
		Value:   translator.Get(key, map[string]any{"name": "Taylor"}),
		Locale:  translator.Locale(),
		Details: []string{"handled_locale:" + handledLocale, "declined:" + translator.Get("messages.still_missing", nil)},
	}
}

func groupPathsResult(translator transcontract.Translator, basePath string) Result {
	path := filepath.Join(basePath, "app", "demo", "translation", "testdata", "group")
	translator.AddPath(path)

	const key = "messages.path_source"
	return Result{
		Case:    "group-paths",
		Key:     key,
		Value:   translator.Get(key, nil),
		Locale:  translator.Locale(),
		Details: []string{"override:" + translator.Get("messages.welcome", nil)},
	}
}

func jsonPathsResult(translator transcontract.Translator, basePath string) Result {
	path := filepath.Join(basePath, "app", "demo", "translation", "testdata", "json")
	translator.AddJSONPath(path)

	const key = "JSON Greeting"
	return Result{
		Case:    "json-paths",
		Key:     key,
		Value:   translator.Get(key, nil),
		Locale:  translator.Locale(),
		Details: []string{"later_path_override:true"},
	}
}

func customLoaderResult(translator transcontract.Translator) (Result, error) {
	concrete, ok := translator.(*frameworktranslation.Translator)
	if !ok {
		return Result{}, fmt.Errorf("translation demo custom-loader: resolved translator has type %T", translator)
	}

	loader := newDemoLoader(map[string]any{"source": "Custom loader value"})
	concrete.SetLoader(loader)
	const key = "messages.source"
	return Result{
		Case:    "custom-loader",
		Key:     key,
		Value:   concrete.Get(key, nil),
		Locale:  concrete.Locale(),
		Details: []string{fmt.Sprintf("loader:%T", concrete.Loader()), "loads:" + strconv.Itoa(loader.loads)},
	}, nil
}

func translatorContractResult() Result {
	var translator transcontract.Translator = frameworktranslation.NewTranslator(frameworktranslation.NewFileLoader(), "en", "fr")
	translator.AddLines(map[string]any{"messages.contract": "Translator contract"}, "en")
	return Result{
		Case:    "translator-contract",
		Key:     "messages.contract",
		Value:   translator.Get("messages.contract", nil),
		Locale:  translator.Locale(),
		Details: []string{"contract:translation.Translator", "fallback:" + translator.GetFallback()},
	}
}

func loaderContractResult(basePath string) Result {
	var loader transcontract.Loader = frameworktranslation.NewFileLoader()
	groupPath := filepath.Join(basePath, "app", "demo", "translation", "testdata", "group")
	jsonPath := filepath.Join(basePath, "app", "demo", "translation", "testdata", "json")
	loader.AddPath(groupPath)
	loader.AddPath(groupPath)
	loader.AddJSONPath(jsonPath)
	loader.AddJSONPath(jsonPath)
	loader.AddNamespace("demo", groupPath)

	return Result{
		Case:   "loader-contract",
		Key:    "translation.Loader",
		Value:  fmt.Sprintf("paths:%d,json_paths:%d", len(loader.Paths()), len(loader.JSONPaths())),
		Locale: "en",
		Details: []string{
			"namespace_registered:" + strconv.FormatBool(loader.Namespaces()["demo"] == groupPath),
			"duplicate_paths_ignored:true",
		},
	}
}

func selectorContractResult() Result {
	var selector transcontract.Selector = frameworktranslation.NewMessageSelector()
	return Result{
		Case:    "selector-contract",
		Key:     "translation.Selector",
		Value:   selector.Select("One result|:count results", 2, "en"),
		Locale:  "en",
		Details: []string{"one:" + selector.Select("One result|:count results", 1, "en")},
	}
}

func resetResult(translator transcontract.Translator) Result {
	// Facades resolve directly from the current Application container, so Reset has
	// no package-level cache to discard and leaves the application singleton intact.
	frameworktranslation.Reset()
	resolved := frameworktranslation.Resolve()
	const key = "messages.welcome"
	return Result{
		Case:    "reset",
		Key:     key,
		Value:   resolved.Get(key, nil),
		Locale:  resolved.Locale(),
		Details: []string{"application_scoped:true", "same_instance:" + strconv.FormatBool(resolved == translator)},
	}
}

func isolatedResult() Result {
	first := frameworktranslation.NewTranslator(frameworktranslation.NewFileLoader(), "en", "en")
	second := frameworktranslation.NewTranslator(frameworktranslation.NewFileLoader(), "en", "en")
	first.AddLines(map[string]any{"messages.scope": "First translator"}, "en")

	const key = "messages.scope"
	return Result{
		Case:    "isolated",
		Key:     key,
		Value:   first.Get(key, nil),
		Locale:  first.Locale(),
		Details: []string{"second:" + second.Get(key, nil), "independent:" + strconv.FormatBool(first.Get(key, nil) != second.Get(key, nil))},
	}
}

type demoLoader struct {
	lines      map[string]any
	namespaces map[string]string
	paths      []string
	jsonPaths  []string
	loads      int
}

func newDemoLoader(lines map[string]any) *demoLoader {
	return &demoLoader{lines: lines, namespaces: make(map[string]string)}
}

func (l *demoLoader) Load(locale, group, namespace string) (map[string]any, error) {
	l.loads++
	if locale != "en" || group != "messages" || namespace != "*" {
		return map[string]any{}, nil
	}
	result := make(map[string]any, len(l.lines))
	for key, value := range l.lines {
		result[key] = value
	}
	return result, nil
}

func (l *demoLoader) AddNamespace(namespace, hint string) {
	l.namespaces[namespace] = hint
}

func (l *demoLoader) AddPath(path string) {
	l.paths = append(l.paths, path)
}

func (l *demoLoader) AddJSONPath(path string) {
	l.jsonPaths = append(l.jsonPaths, path)
}

func (l *demoLoader) Namespaces() map[string]string {
	result := make(map[string]string, len(l.namespaces))
	for namespace, hint := range l.namespaces {
		result[namespace] = hint
	}
	return result
}

func (l *demoLoader) Paths() []string {
	return append([]string(nil), l.paths...)
}

func (l *demoLoader) JSONPaths() []string {
	return append([]string(nil), l.jsonPaths...)
}

func pluralReplacementsResult(translator transcontract.Translator) Result {
	const key = "messages.minutes_ago"
	return translatedResult(
		"plural-replacements",
		key,
		translator.Choice(key, 5, map[string]any{"value": 5}),
		translator.Locale(),
	)
}

func pluralCountResult(translator transcontract.Translator) Result {
	const key = "messages.notifications"
	return Result{
		Case:    "plural-count",
		Key:     key,
		Value:   translator.Choice(key, 10, nil),
		Locale:  translator.Locale(),
		Details: []string{"custom_count:" + translator.Choice(key, 10, map[string]any{"count": "many"})},
	}
}

func localeResult(translator transcontract.Translator) (Result, error) {
	previous := translator.CurrentLocale()
	if err := translator.SetLocale("fr"); err != nil {
		return Result{}, fmt.Errorf("translation demo locale set current locale: %w", err)
	}

	const key = "messages.welcome"
	return Result{
		Case:   "locale",
		Key:    key,
		Value:  translator.Get(key, nil),
		Locale: translator.Locale(),
		Details: []string{
			"previous:" + previous,
			"current:" + translator.CurrentLocale(),
			"is_fr:" + strconv.FormatBool(translator.IsLocale("fr")),
		},
	}, nil
}

func localeValidationResult(translator transcontract.Translator) Result {
	previous := translator.Locale()
	err := translator.SetLocale("en/../private")
	return Result{
		Case:   "locale-validation",
		Key:    "locale",
		Value:  strconv.FormatBool(err != nil),
		Locale: translator.Locale(),
		Details: []string{
			"unchanged:" + strconv.FormatBool(translator.Locale() == previous),
			"error:" + errorText(err),
		},
	}
}

func fallbackResult(translator transcontract.Translator) (Result, error) {
	previous := translator.GetFallback()
	target := "en"
	key := "messages.english_only"
	if previous == target {
		target = "fr"
		key = "messages.welcome"
	}
	if err := translator.SetLocale("es"); err != nil {
		return Result{}, fmt.Errorf("translation demo fallback set current locale: %w", err)
	}
	if err := translator.SetFallback(target); err != nil {
		return Result{}, fmt.Errorf("translation demo fallback set fallback locale: %w", err)
	}

	return Result{
		Case:   "fallback",
		Key:    key,
		Value:  translator.Get(key, nil),
		Locale: translator.Locale(),
		Details: []string{
			"previous:" + previous,
			"fallback:" + translator.GetFallback(),
		},
	}, nil
}

func fallbackValidationResult(translator transcontract.Translator) Result {
	previous := translator.GetFallback()
	err := translator.SetFallback(`fr\..\private`)
	return Result{
		Case:   "fallback-validation",
		Key:    "fallback_locale",
		Value:  strconv.FormatBool(err != nil),
		Locale: translator.Locale(),
		Details: []string{
			"unchanged:" + strconv.FormatBool(translator.GetFallback() == previous),
			"fallback:" + translator.GetFallback(),
			"error:" + errorText(err),
		},
	}
}

func localeResolverResult(translator transcontract.Translator) Result {
	var requestedLocale string
	translator.DetermineLocalesUsing(func(_ string, requested string) []string {
		requestedLocale = requested
		return []string{"fr", "en"}
	})

	const key = "messages.welcome"
	return Result{
		Case:    "locale-resolver",
		Key:     key,
		Value:   translator.Get(key, nil),
		Locale:  translator.Locale(),
		Details: []string{"requested:" + requestedLocale, "chain:fr,en"},
	}
}

func namespaceOverridesResult(translator transcontract.Translator, basePath string) Result {
	translator.AddNamespace("hearthfire", filepath.Join(basePath, "package-lang", "hearthfire"))
	const key = "hearthfire::messages.fire"
	return Result{
		Case:    "namespace-overrides",
		Key:     key,
		Value:   translator.Get(key, nil),
		Locale:  translator.Locale(),
		Details: []string{"package_only:" + translator.Get("hearthfire::messages.package_only", nil)},
	}
}

func addLinesResult(translator transcontract.Translator) Result {
	translator.AddLines(map[string]any{
		"messages.runtime": "Hello, :Name",
	}, "en")
	translator.AddLines(map[string]any{
		"alerts.info": "Info from acme",
	}, "en", "acme")

	const key = "messages.runtime"
	return Result{
		Case:    "add-lines",
		Key:     key,
		Value:   translator.Get(key, map[string]any{"name": "Taylor"}),
		Locale:  translator.Locale(),
		Details: []string{"namespaced:" + translator.Get("acme::alerts.info", nil)},
	}
}

func addLinesPrecedenceResult(translator transcontract.Translator) Result {
	const key = "messages.welcome"
	fileValue := translator.Get(key, nil)
	translator.AddLines(map[string]any{key: "Runtime welcome"}, "en")
	return Result{
		Case:    "add-lines-precedence",
		Key:     key,
		Value:   translator.Get(key, nil),
		Locale:  translator.Locale(),
		Details: []string{"file:" + fileValue},
	}
}

func errorText(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

func localeArgumentResult(translator transcontract.Translator) Result {
	const key = "messages.welcome"
	return Result{
		Case:    "locale-argument",
		Key:     key,
		Value:   translator.Get(key, nil, "fr"),
		Locale:  translator.Locale(),
		Details: []string{"requested_locale:fr", "current_locale_unchanged:" + translator.Locale()},
	}
}

func hasResult(translator transcontract.Translator) Result {
	const key = "messages.welcome"
	return Result{
		Case:   "has",
		Key:    key,
		Value:  strconv.FormatBool(translator.Has(key)),
		Locale: translator.Locale(),
		Details: []string{
			"missing:" + strconv.FormatBool(translator.Has("messages.missing")),
			"group:" + strconv.FormatBool(translator.Has("messages")),
		},
	}
}

func hasForLocaleResult(translator transcontract.Translator) Result {
	const key = "messages.welcome"
	return Result{
		Case:   "has-for-locale",
		Key:    key,
		Value:  strconv.FormatBool(translator.HasForLocale(key, "fr")),
		Locale: translator.Locale(),
		Details: []string{
			"fr:" + strconv.FormatBool(translator.HasForLocale(key, "fr")),
			"fr_english_only:" + strconv.FormatBool(translator.HasForLocale("messages.english_only", "fr")),
		},
	}
}

func getMapResult(translator transcontract.Translator) (Result, error) {
	messages := translator.GetMap("messages")
	if messages == nil {
		return Result{}, fmt.Errorf("translation demo get-map: messages is missing")
	}
	auth := translator.GetMap("messages.auth")
	if auth == nil {
		return Result{}, fmt.Errorf("translation demo get-map: messages.auth is missing")
	}
	welcome, welcomeOK := messages["welcome"].(string)
	failed, failedOK := auth["failed"].(string)
	if !welcomeOK || !failedOK {
		return Result{}, fmt.Errorf("translation demo get-map: expected string values are missing")
	}
	return Result{
		Case:    "get-map",
		Key:     "messages",
		Value:   welcome,
		Locale:  translator.Locale(),
		Details: []string{"auth.failed:" + failed},
	}, nil
}

type demoPrice struct {
	amount float64
}

type demoPerson string

func (p demoPerson) String() string {
	return string(p)
}

func stringableResult(translator transcontract.Translator) Result {
	translator.Stringable(demoPrice{}, func(value any) string {
		price := value.(demoPrice)
		return fmt.Sprintf("$%.2f", price.amount)
	})

	const key = "messages.receipt"
	return Result{
		Case: "stringable",
		Key:  key,
		Value: translator.Get(key, map[string]any{
			"price": demoPrice{amount: 12.5},
			"name":  demoPerson("Ada"),
		}),
		Locale:  translator.Locale(),
		Details: []string{"custom_formatter:demoPrice", "fmt.Stringer:demoPerson"},
	}
}

func pluralPipeResult(translator transcontract.Translator) Result {
	const key = "messages.apples"
	return Result{
		Case:    "plural-pipe",
		Key:     key,
		Value:   translator.Choice(key, 3, nil),
		Locale:  translator.Locale(),
		Details: []string{"one:" + translator.Choice(key, 1, nil)},
	}
}

func pluralIntervalsResult(translator transcontract.Translator) Result {
	const key = "messages.invitations"
	return Result{
		Case:   "plural-intervals",
		Key:    key,
		Value:  translator.Choice(key, 0, nil),
		Locale: translator.Locale(),
		Details: []string{
			"one:" + translator.Choice(key, 1, nil),
			"many:" + translator.Choice(key, 5, nil),
		},
	}
}

func architectureResult(translator transcontract.Translator) Result {
	return Result{
		Case:   "architecture",
		Key:    "translator",
		Value:  fmt.Sprintf("%T", translator),
		Locale: translator.Locale(),
		Details: []string{
			"contract:translation.Translator",
			fmt.Sprintf("loader:%T", frameworktranslation.Loader()),
		},
	}
}

func configResult(translator transcontract.Translator) Result {
	return Result{
		Case:    "config",
		Key:     "app.locale",
		Value:   translator.Locale(),
		Locale:  translator.Locale(),
		Details: []string{"app.fallback_locale:" + translator.GetFallback()},
	}
}

func providerResult(translator transcontract.Translator) Result {
	resolvedAgain := frameworktranslation.Resolve()
	loader := frameworktranslation.Loader()
	loaderAgain := frameworktranslation.Loader()
	return Result{
		Case:   "provider",
		Key:    "translator",
		Value:  "registered",
		Locale: translator.Locale(),
		Details: []string{
			fmt.Sprintf("translator_singleton:%t", translator == resolvedAgain),
			fmt.Sprintf("loader_registered:%t", loader != nil),
			fmt.Sprintf("loader_singleton:%t", loader != nil && loader == loaderAgain),
		},
	}
}

func nestedKeysResult(translator transcontract.Translator) (Result, error) {
	auth := translator.GetMap("messages.auth", translator.Locale())
	if auth == nil {
		return Result{}, fmt.Errorf("translation demo nested-keys: messages.auth is missing")
	}
	failed, ok := auth["failed"].(string)
	if !ok {
		return Result{}, fmt.Errorf("translation demo nested-keys: messages.auth.failed is not a string")
	}
	return translatedResult("nested-keys", "messages.auth.failed", failed, translator.Locale()), nil
}

func keyConflictResult(translator transcontract.Translator) Result {
	const key = "action"
	group := translator.GetMap(key, translator.Locale())
	groupTitle, _ := group["title"].(string)
	return Result{
		Case:    "key-conflicts",
		Key:     key,
		Value:   translator.Get(key, nil),
		Locale:  translator.Locale(),
		Details: []string{fmt.Sprintf("has:%t", translator.Has(key)), "group.title:" + groupTitle},
	}
}

func namespaceResult(translator transcontract.Translator, basePath string) Result {
	translator.AddNamespace("acme", filepath.Join(basePath, "lang", "vendor", "acme"))
	const key = "acme::messages.hello"
	return translatedResult("namespaces", key, translator.Get(key, nil), translator.Locale())
}

func translatedResult(caseName, key, value, locale string) Result {
	return Result{Case: caseName, Key: key, Value: value, Locale: locale}
}
