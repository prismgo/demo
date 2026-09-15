// Package routedemo contains runnable examples of the Laravel-style HTTP router.
package routedemo

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// Result records one observable route scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a route catalog scenario.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("route demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

// run isolates Gin's process-wide mode and writer while a scenario executes.
func run(name string) (string, error) {
	mode := gin.Mode()
	writer := gin.DefaultWriter
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	defer func() {
		gin.SetMode(mode)
		gin.DefaultWriter = writer
	}()

	switch name {
	case "architecture":
		return architectureScenario()
	case "facade":
		return facadeScenario()
	case "mount":
		return mountScenario()
	case "isolated-router":
		return isolatedRouterScenario()
	case "http-methods":
		return httpMethodsScenario()
	case "match":
		return matchScenario()
	case "any":
		return anyScenario()
	case "laravel-parameters":
		return laravelParametersScenario()
	case "gin-parameters":
		return ginParametersScenario()
	case "parameter-read":
		return parameterReadScenario()
	case "optional-parameters":
		return optionalParametersScenario()
	case "wildcard-parameters":
		return wildcardParametersScenario()
	case "where":
		return whereScenario()
	case "where-number":
		return whereNumberScenario()
	case "where-alpha":
		return whereAlphaScenario()
	case "where-alphanumeric":
		return whereAlphaNumericScenario()
	case "where-uuid":
		return whereUUIDScenario()
	case "where-ulid":
		return whereULIDScenario()
	case "where-in":
		return whereInScenario()
	case "group-constraints":
		return groupConstraintsScenario()
	case "global-pattern":
		return globalPatternScenario()
	case "constraint-override":
		return constraintOverrideScenario()
	case "route-name":
		return routeNameScenario()
	case "url":
		return urlScenario()
	case "url-escaping":
		return urlEscapingScenario()
	case "url-missing-parameter":
		return urlMissingParameterScenario()
	case "group-name-prefix":
		return groupNamePrefixScenario()
	case "duplicate-route-name":
		return duplicateRouteNameScenario()
	case "group-prefix":
		return groupPrefixScenario()
	case "group-chain":
		return groupChainScenario()
	case "nested-groups":
		return nestedGroupsScenario()
	case "group-middleware":
		return groupMiddlewareScenario()
	case "route-middleware":
		return routeMiddlewareScenario()
	case "named-middleware":
		return namedMiddlewareScenario()
	case "middleware-function-name":
		return middlewareFunctionNameScenario()
	case "route-without-middleware":
		return routeWithoutMiddlewareScenario()
	case "registrar-without-middleware":
		return registrarWithoutMiddlewareScenario()
	case "bind":
		return bindScenario()
	case "model":
		return modelScenario()
	case "binding-context":
		return bindingContextScenario()
	case "missing-handler":
		return missingHandlerScenario()
	case "default-missing":
		return defaultMissingScenario()
	case "controller-action":
		return controllerActionScenario()
	case "controller-validation":
		return controllerValidationScenario()
	case "api-resource":
		return apiResourceScenario()
	case "resource":
		return resourceScenario()
	case "resource-create":
		return resourceCreateScenario()
	case "resource-edit":
		return resourceEditScenario()
	case "resource-only":
		return resourceOnlyScenario()
	case "resource-except":
		return resourceExceptScenario()
	case "resource-names":
		return resourceNamesScenario()
	case "resource-parameters":
		return resourceParametersScenario()
	case "api-resources":
		return apiResourcesScenario()
	case "nested-resource":
		return nestedResourceScenario()
	case "resource-controller-contract":
		return resourceControllerContractScenario()
	case "create-controller-contract":
		return createControllerContractScenario()
	case "edit-controller-contract":
		return editControllerContractScenario()
	case "redirect":
		return redirectScenario()
	case "permanent-redirect":
		return permanentRedirectScenario()
	case "static":
		return staticScenario()
	case "fallback":
		return fallbackScenario()
	case "domain":
		return domainScenario()
	case "domain-placeholder":
		return domainPlaceholderScenario()
	case "rate-limiter":
		return rateLimiterScenario()
	case "limit":
		return limitScenario()
	case "throttle-route":
		return throttleRouteScenario()
	case "throttle-group":
		return throttleGroupScenario()
	case "throttle-unknown":
		return throttleUnknownScenario()
	case "throttle-over-limit":
		return throttleOverLimitScenario()
	case "current-route":
		return currentRouteScenario()
	case "route-info":
		return routeInfoScenario()
	case "list":
		return listScenario()
	case "list-command":
		return listCommandScenario()
	case "handler-order":
		return handlerOrderScenario()
	case "resolve":
		return resolveScenario()
	case "reset":
		return resetScenario()
	case "clone":
		return cloneScenario()
	case "add":
		return addScenario()
	case "router-group":
		return routerGroupScenario()
	case "route-scope-bindings":
		return routeScopeBindingsScenario()
	case "registrar-scope-bindings":
		return registrarScopeBindingsScenario()
	case "registrar-overrides":
		return registrarOverridesScenario()
	case "facade-contract":
		return facadeContractScenario()
	case "provider-registration":
		return providerRegistrationScenario()
	case "provider-preserves-router":
		return providerPreservesRouterScenario()
	case "provider-singleton":
		return providerSingletonScenario()
	case "best-practices":
		return bestPracticesScenario()
	default:
		return "", fmt.Errorf("unknown scenario %q", name)
	}
}

// architectureScenario keeps the router contracts referenced at compile time.
func architectureScenario() (string, error) {
	var (
		_ func() *route.Router                                                     = route.New
		_ func(*route.Router, string, ...route.HandlerFunc) *route.Route           = (*route.Router).Get
		_ func(*route.Router, []string, string, ...route.HandlerFunc) *route.Route = (*route.Router).Match
		_ func(*route.Router, *gin.Engine) error                                   = (*route.Router).Mount
		_ func(*route.Router) []route.RouteInfo                                    = (*route.Router).List
		_ func(*route.Router, string, map[string]any) (string, error)              = (*route.Router).URL
		_ func(*route.Route, string, string) *route.Route                          = (*route.Route).Where
		_ func(*route.Route, string) *route.Route                                  = (*route.Route).WhereNumber
		_ route.HandlerFunc                                                        = func(*gin.Context) {}
		_ route.Binder                                                             = func(*gin.Context, string) (any, error) { return nil, nil }
		_ route.RouteInfo
	)
	return "router=Router route=Route info=RouteInfo binder=Binder mount=Mount", nil
}

// facadeScenario drives the global Router resolved from the running Application.
func facadeScenario() (string, error) {
	router := route.Resolve()
	router.Reset()
	defer router.Reset()

	route.Get("/facade/users/{id}", func(c *gin.Context) {
		c.String(http.StatusOK, "user="+c.Param("id"))
	}).Name("facade.users.show")

	recorder, err := dispatch(router, http.MethodGet, "/facade/users/7")
	if err != nil {
		return "", err
	}
	generated, err := route.URL("facade.users.show", map[string]any{"id": 7})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("status=%d body=%s url=%s singleton=%t",
		recorder.Code, recorder.Body.String(), generated, route.Resolve() == router), nil
}

// mountScenario mounts one router onto a fresh Gin engine.
func mountScenario() (string, error) {
	router := route.New()
	router.Get("/ping", textHandler("pong"))
	recorder, err := dispatch(router, http.MethodGet, "/ping")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("status=%d body=%s routes=%d",
		recorder.Code, recorder.Body.String(), len(router.List())), nil
}

// isolatedRouterScenario proves independent routers keep separate route tables.
func isolatedRouterScenario() (string, error) {
	first := route.New()
	first.Get("/one", textHandler("one"))
	second := route.New()
	second.Get("/two", textHandler("two"))

	firstOne, err := statusOf(first, http.MethodGet, "/one")
	if err != nil {
		return "", err
	}
	firstTwo, err := statusOf(first, http.MethodGet, "/two")
	if err != nil {
		return "", err
	}
	secondTwo, err := statusOf(second, http.MethodGet, "/two")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("first-one=%d first-two=%d second-two=%d", firstOne, firstTwo, secondTwo), nil
}

// httpMethodsScenario registers one handler per documented HTTP method.
func httpMethodsScenario() (string, error) {
	router := route.New()
	handler := func(c *gin.Context) { c.String(http.StatusOK, c.Request.Method) }
	router.Get("/method", handler)
	router.Post("/method", handler)
	router.Put("/method", handler)
	router.Patch("/method", handler)
	router.Delete("/method", handler)
	router.Options("/method", handler)

	parts := make([]string, 0, 6)
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions} {
		recorder, err := dispatch(router, method, "/method")
		if err != nil {
			return "", err
		}
		parts = append(parts, strings.ToLower(method)+"="+recorder.Body.String())
	}
	return strings.Join(parts, " "), nil
}

// matchScenario registers one handler for a method set.
func matchScenario() (string, error) {
	router := route.New()
	router.Match([]string{http.MethodPut, http.MethodPatch}, "/users/{id}", paramHandler("id"))

	put, err := bodyOf(router, http.MethodPut, "/users/42")
	if err != nil {
		return "", err
	}
	patch, err := bodyOf(router, http.MethodPatch, "/users/42")
	if err != nil {
		return "", err
	}
	get, err := statusOf(router, http.MethodGet, "/users/42")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("put=%s patch=%s get=%d", put, patch, get), nil
}

// anyScenario registers every common HTTP method on one URI.
func anyScenario() (string, error) {
	router := route.New()
	router.Any("/webhook", func(c *gin.Context) { c.String(http.StatusOK, c.Request.Method) })

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead}
	statuses := make([]string, 0, len(methods))
	for _, method := range methods {
		status, err := statusOf(router, method, "/webhook")
		if err != nil {
			return "", err
		}
		statuses = append(statuses, fmt.Sprint(status))
	}
	return fmt.Sprintf("methods=%d statuses=%s", len(methods), strings.Join(statuses, ",")), nil
}

// laravelParametersScenario reads parameters declared with the {name} syntax.
func laravelParametersScenario() (string, error) {
	router := route.New()
	router.Get("/users/{id}", paramHandler("id"))
	router.Get("/posts/{slug}", paramHandler("slug"))

	id, err := bodyOf(router, http.MethodGet, "/users/100")
	if err != nil {
		return "", err
	}
	slug, err := bodyOf(router, http.MethodGet, "/posts/hello")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("id=%s slug=%s", id, slug), nil
}

// ginParametersScenario reads parameters declared with the native :name syntax.
func ginParametersScenario() (string, error) {
	router := route.New()
	router.Get("/legacy/:id", paramHandler("id"))
	value, err := bodyOf(router, http.MethodGet, "/legacy/7")
	if err != nil {
		return "", err
	}
	return "id=" + value, nil
}

// parameterReadScenario reads multiple parameters inside one handler.
func parameterReadScenario() (string, error) {
	router := route.New()
	router.Get("/teams/{team}/members/{member}", func(c *gin.Context) {
		c.String(http.StatusOK, c.Param("team")+"/"+c.Param("member"))
	})
	value, err := bodyOf(router, http.MethodGet, "/teams/alpha/members/beta")
	if err != nil {
		return "", err
	}
	return "params=" + value, nil
}

// optionalParametersScenario exercises the trailing {name?} form.
func optionalParametersScenario() (string, error) {
	router := route.New()
	router.Get("/posts/{slug?}", func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" {
			slug = "none"
		}
		c.String(http.StatusOK, slug)
	})
	with, err := bodyOf(router, http.MethodGet, "/posts/hello")
	if err != nil {
		return "", err
	}
	without, err := bodyOf(router, http.MethodGet, "/posts")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("with=%s without=%s", with, without), nil
}

// wildcardParametersScenario compiles a trailing .* constraint into a Gin catch-all.
func wildcardParametersScenario() (string, error) {
	router := route.New()
	router.Get("/files/{path}", func(c *gin.Context) {
		c.String(http.StatusOK, strings.TrimPrefix(c.Param("path"), "/"))
	}).Where("path", ".*")
	value, err := bodyOf(router, http.MethodGet, "/files/assets/js/app.js")
	if err != nil {
		return "", err
	}
	return "path=" + value, nil
}

// whereScenario applies a custom regular-expression constraint.
func whereScenario() (string, error) {
	router := route.New()
	router.Get("/codes/{code}", textHandler("match")).Where("code", "[A-Z]{2}[0-9]{3}")
	valid, err := statusOf(router, http.MethodGet, "/codes/AB123")
	if err != nil {
		return "", err
	}
	invalid, err := statusOf(router, http.MethodGet, "/codes/ab123")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("valid=%d invalid=%d", valid, invalid), nil
}

// whereNumberScenario applies the numeric constraint helper.
func whereNumberScenario() (string, error) {
	router := route.New()
	router.Get("/numbers/{id}", textHandler("number")).WhereNumber("id")
	valid, err := statusOf(router, http.MethodGet, "/numbers/42")
	if err != nil {
		return "", err
	}
	alpha, err := statusOf(router, http.MethodGet, "/numbers/abc")
	if err != nil {
		return "", err
	}
	negative, err := statusOf(router, http.MethodGet, "/numbers/-1")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("valid=%d alpha=%d negative=%d", valid, alpha, negative), nil
}

// whereAlphaScenario applies the alphabetic constraint helper.
func whereAlphaScenario() (string, error) {
	router := route.New()
	router.Get("/letters/{name}", textHandler("alpha")).WhereAlpha("name")
	valid, err := statusOf(router, http.MethodGet, "/letters/abc")
	if err != nil {
		return "", err
	}
	numeric, err := statusOf(router, http.MethodGet, "/letters/ab1")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("valid=%d numeric=%d", valid, numeric), nil
}

// whereAlphaNumericScenario applies the alphanumeric constraint helper.
func whereAlphaNumericScenario() (string, error) {
	router := route.New()
	router.Get("/codes/{code}", textHandler("alnum")).WhereAlphaNumeric("code")
	valid, err := statusOf(router, http.MethodGet, "/codes/AB12")
	if err != nil {
		return "", err
	}
	symbol, err := statusOf(router, http.MethodGet, "/codes/ab-1")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("valid=%d symbol=%d", valid, symbol), nil
}

// whereUUIDScenario applies the UUID constraint helper.
func whereUUIDScenario() (string, error) {
	router := route.New()
	router.Get("/orders/{uuid}", textHandler("uuid")).WhereUuid("uuid")
	valid, err := statusOf(router, http.MethodGet, "/orders/550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		return "", err
	}
	invalid, err := statusOf(router, http.MethodGet, "/orders/not-a-uuid")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("valid=%d invalid=%d", valid, invalid), nil
}

// whereULIDScenario applies the ULID constraint helper.
func whereULIDScenario() (string, error) {
	router := route.New()
	router.Get("/events/{ulid}", textHandler("ulid")).WhereUlid("ulid")
	valid, err := statusOf(router, http.MethodGet, "/events/01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		return "", err
	}
	invalid, err := statusOf(router, http.MethodGet, "/events/01arz3ndektsv4rrffq69g5fav")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("valid=%d invalid=%d", valid, invalid), nil
}

// whereInScenario applies the enumerated constraint helper.
func whereInScenario() (string, error) {
	router := route.New()
	router.Get("/status/{value}", textHandler("status")).WhereIn("value", []string{"open", "closed"})
	open, err := statusOf(router, http.MethodGet, "/status/open")
	if err != nil {
		return "", err
	}
	closed, err := statusOf(router, http.MethodGet, "/status/closed")
	if err != nil {
		return "", err
	}
	other, err := statusOf(router, http.MethodGet, "/status/pending")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("open=%d closed=%d other=%d", open, closed, other), nil
}

// groupConstraintsScenario applies one constraint across every group route.
func groupConstraintsScenario() (string, error) {
	router := route.New()
	router.Prefix("/api").Where("id", `\d+`).Group(func() {
		router.Get("/users/{id}", textHandler("user"))
		router.Get("/orders/{id}", textHandler("order"))
	})
	users, err := statusOf(router, http.MethodGet, "/api/users/5")
	if err != nil {
		return "", err
	}
	orders, err := statusOf(router, http.MethodGet, "/api/orders/5")
	if err != nil {
		return "", err
	}
	invalid, err := statusOf(router, http.MethodGet, "/api/users/x")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("users=%d orders=%d invalid=%d", users, orders, invalid), nil
}

// globalPatternScenario applies a router-wide pattern to route parameters.
func globalPatternScenario() (string, error) {
	router := route.New()
	router.Pattern("slug", `[a-z0-9-]+`)
	router.Get("/posts/{slug}", textHandler("post"))
	valid, err := statusOf(router, http.MethodGet, "/posts/hello-world")
	if err != nil {
		return "", err
	}
	invalid, err := statusOf(router, http.MethodGet, "/posts/Hello")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("valid=%d invalid=%d", valid, invalid), nil
}

// constraintOverrideScenario proves a route Where beats the global pattern.
func constraintOverrideScenario() (string, error) {
	router := route.New()
	router.Pattern("id", `\d+`)
	router.Get("/pattern/{id}", textHandler("pattern"))
	router.Get("/local/{id}", textHandler("local")).Where("id", `[A-Z]+`)
	patternValid, err := statusOf(router, http.MethodGet, "/pattern/42")
	if err != nil {
		return "", err
	}
	patternInvalid, err := statusOf(router, http.MethodGet, "/pattern/abc")
	if err != nil {
		return "", err
	}
	localValid, err := statusOf(router, http.MethodGet, "/local/ABC")
	if err != nil {
		return "", err
	}
	localInvalid, err := statusOf(router, http.MethodGet, "/local/42")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pattern=%d,%d local=%d,%d", patternValid, patternInvalid, localValid, localInvalid), nil
}

// routeNameScenario records a route name in the route snapshot.
func routeNameScenario() (string, error) {
	router := route.New()
	router.Get("/users/{id}", textHandler("user")).Name("users.show")
	infos := router.List()
	name := ""
	for _, info := range infos {
		if info.Name != "" {
			name = info.Name
		}
	}
	return fmt.Sprintf("name=%s entries=%d", name, len(infos)), nil
}

// urlScenario generates a path from a named route.
func urlScenario() (string, error) {
	router := route.New()
	router.Get("/users/{id}", textHandler("user")).Name("users.show")
	path, err := router.URL("users.show", map[string]any{"id": 100})
	if err != nil {
		return "", err
	}
	return "url=" + path, nil
}

// urlEscapingScenario escapes parameter values with url.PathEscape.
func urlEscapingScenario() (string, error) {
	router := route.New()
	router.Get("/search/{term}", textHandler("search")).Name("search")
	path, err := router.URL("search", map[string]any{"term": "hello world/x"})
	if err != nil {
		return "", err
	}
	return "url=" + path, nil
}

// urlMissingParameterScenario reports an error when a required parameter is absent.
func urlMissingParameterScenario() (string, error) {
	router := route.New()
	router.Get("/users/{id}", textHandler("user")).Name("users.show")
	_, err := router.URL("users.show", nil)
	return fmt.Sprintf("missing=%t", err != nil), nil
}

// groupNamePrefixScenario combines a path prefix with a name prefix.
func groupNamePrefixScenario() (string, error) {
	router := route.New()
	router.Prefix("/admin").Name("admin.").Group(func() {
		router.Get("/users/{id}", textHandler("user")).Name("users.show")
	})
	path, err := router.URL("admin.users.show", map[string]any{"id": 100})
	if err != nil {
		return "", err
	}
	return "url=" + path, nil
}

// duplicateRouteNameScenario replaces a route name suffix while keeping the prefix.
func duplicateRouteNameScenario() (string, error) {
	router := route.New()
	ref := router.Name("admin.").Get("/users/{id}", textHandler("user"))
	ref.Name("users.show")
	ref.Name("users.detail")
	detail, err := router.URL("admin.users.detail", map[string]any{"id": 100})
	if err != nil {
		return "", err
	}
	_, oldErr := router.URL("admin.users.show", map[string]any{"id": 100})
	return fmt.Sprintf("detail=%s old-missing=%t", detail, oldErr != nil), nil
}

// groupPrefixScenario shares one path prefix across group routes.
func groupPrefixScenario() (string, error) {
	router := route.New()
	router.Prefix("/api/v1").Group(func() {
		router.Get("/users", textHandler("index"))
		router.Get("/users/{id}", textHandler("show"))
	})
	index, err := statusOf(router, http.MethodGet, "/api/v1/users")
	if err != nil {
		return "", err
	}
	show, err := statusOf(router, http.MethodGet, "/api/v1/users/9")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("index=%d show=%d", index, show), nil
}

// groupChainScenario chains path, name, and middleware attributes.
func groupChainScenario() (string, error) {
	router := route.New()
	chain := func(c *gin.Context) {
		c.Set("chain", "yes")
		c.Header("X-Demo-Chain", "yes")
		c.Next()
	}
	router.Prefix("/api/v1").Name("api.").Middleware(chain).Group(func() {
		router.Get("/profile", func(c *gin.Context) {
			c.String(http.StatusOK, c.GetString("chain"))
		}).Name("profile")
	})
	recorder, err := dispatch(router, http.MethodGet, "/api/v1/profile")
	if err != nil {
		return "", err
	}
	name := ""
	for _, info := range router.List() {
		if info.Name != "" {
			name = info.Name
		}
	}
	return fmt.Sprintf("name=%s value=%s header=%s", name, recorder.Body.String(), recorder.Header().Get("X-Demo-Chain")), nil
}

// nestedGroupsScenario inherits outer attributes in an inner group.
func nestedGroupsScenario() (string, error) {
	router := route.New()
	router.Prefix("/api").Name("api.").Group(func() {
		router.Prefix("/admin").Name("admin.").Group(func() {
			router.Get("/users", textHandler("users")).Name("users.index")
		})
	})
	path, err := router.URL("api.admin.users.index", nil)
	if err != nil {
		return "", err
	}
	name := ""
	for _, info := range router.List() {
		if info.Name != "" {
			name = info.Name
		}
	}
	return fmt.Sprintf("url=%s name=%s", path, name), nil
}

// groupMiddlewareScenario runs a group middleware only for group routes.
func groupMiddlewareScenario() (string, error) {
	router := route.New()
	count := 0
	track := func(c *gin.Context) {
		count++
		c.Next()
	}
	router.Middleware(track).Group(func() {
		router.Get("/first", textHandler("first"))
		router.Get("/second", textHandler("second"))
	})
	router.Get("/outside", textHandler("outside"))
	if _, err := dispatch(router, http.MethodGet, "/first"); err != nil {
		return "", err
	}
	if _, err := dispatch(router, http.MethodGet, "/second"); err != nil {
		return "", err
	}
	groupCount := count
	if _, err := dispatch(router, http.MethodGet, "/outside"); err != nil {
		return "", err
	}
	return fmt.Sprintf("group=%d total=%d", groupCount, count), nil
}

// routeMiddlewareScenario orders route middleware before the action.
func routeMiddlewareScenario() (string, error) {
	router := route.New()
	trace := make([]string, 0, 3)
	track := func(c *gin.Context) {
		trace = append(trace, "middleware")
		c.Next()
	}
	router.Get("/with", track, func(c *gin.Context) {
		trace = append(trace, "action")
		c.String(http.StatusOK, "with")
	})
	router.Get("/bare", func(c *gin.Context) {
		trace = append(trace, "bare")
		c.String(http.StatusOK, "bare")
	})
	if _, err := dispatch(router, http.MethodGet, "/with"); err != nil {
		return "", err
	}
	withOrder := strings.Join(trace, ">")
	trace = trace[:0]
	if _, err := dispatch(router, http.MethodGet, "/bare"); err != nil {
		return "", err
	}
	return fmt.Sprintf("with=%s bare=%s", withOrder, strings.Join(trace, ">")), nil
}

// dispatch mounts the router on a fresh engine and performs one request.
func dispatch(router *route.Router, method, target string) (*httptest.ResponseRecorder, error) {
	return dispatchHost(router, method, target, "")
}

// dispatchHost performs one request while overriding the request Host header.
func dispatchHost(router *route.Router, method, target, host string) (*httptest.ResponseRecorder, error) {
	engine := gin.New()
	if err := router.Mount(engine); err != nil {
		return nil, fmt.Errorf("mount router: %w", err)
	}
	request := httptest.NewRequest(method, target, nil)
	if host != "" {
		request.Host = host
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder, nil
}

// statusOf returns only the response status for one request.
func statusOf(router *route.Router, method, target string) (int, error) {
	recorder, err := dispatch(router, method, target)
	if err != nil {
		return 0, err
	}
	return recorder.Code, nil
}

// bodyOf returns only the response body for one request.
func bodyOf(router *route.Router, method, target string) (string, error) {
	recorder, err := dispatch(router, method, target)
	if err != nil {
		return "", err
	}
	return recorder.Body.String(), nil
}

// textHandler responds with a fixed body.
func textHandler(body string) route.HandlerFunc {
	return func(c *gin.Context) { c.String(http.StatusOK, body) }
}

// paramHandler responds with the value of one path parameter.
func paramHandler(param string) route.HandlerFunc {
	return func(c *gin.Context) { c.String(http.StatusOK, c.Param(param)) }
}
