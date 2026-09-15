package routedemo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// resolveScenario proves the facade always resolves the same global Router singleton.
func resolveScenario() (string, error) {
	router := route.Resolve()
	router.Reset()
	defer router.Reset()

	route.Get("/resolved", textHandler("ok")).Name("resolved")
	return fmt.Sprintf("singleton=%t routes=%d", route.Resolve() == router, len(route.List())), nil
}

// resetScenario clears routes, names and global patterns in one call.
func resetScenario() (string, error) {
	router := route.New()
	router.Pattern("id", `\d+`)
	router.Get("/users/{id}", textHandler("user")).Name("users.show")

	before := len(router.List())
	router.Reset()
	after := len(router.List())
	_, nameErr := router.URL("users.show", map[string]any{"id": 1})

	router.Get("/users/{id}", textHandler("fresh"))
	pattern, err := statusOf(router, http.MethodGet, "/users/abc")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("before=%d after=%d name-missing=%t pattern=%d", before, after, nameErr != nil, pattern), nil
}

// cloneScenario shows that a cloned Router grows a deep-copied route table.
func cloneScenario() (string, error) {
	router := route.New()
	router.Get("/original", textHandler("original")).Name("original")

	cloned := router.Clone()
	cloned.Get("/cloned", textHandler("cloned")).Name("cloned")

	_, originalErr := router.URL("cloned", nil)
	clonedURL, clonedErr := cloned.URL("cloned", nil)
	if clonedErr != nil {
		return "", clonedErr
	}
	return fmt.Sprintf("original=%d cloned=%d isolated=%t url=%s",
		len(router.List()), len(cloned.List()), originalErr != nil, clonedURL), nil
}

// addScenario uses the low-level Add entry to register one handler for several methods.
func addScenario() (string, error) {
	router := route.New()
	router.Add([]string{http.MethodGet, http.MethodPost}, "/logs", textHandler("logs"))

	methods := ""
	for _, info := range router.List() {
		if info.URI == "/logs" {
			methods = strings.Join(info.Methods, ",")
		}
	}
	get, err := statusOf(router, http.MethodGet, "/logs")
	if err != nil {
		return "", err
	}
	post, err := statusOf(router, http.MethodPost, "/logs")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("methods=%s get=%d post=%d", methods, get, post), nil
}

// routerGroupScenario runs a closure directly through Router.Group.
func routerGroupScenario() (string, error) {
	router := route.New()
	router.Group(func() {
		router.Get("/inside", textHandler("inside"))
	})
	router.Get("/outside", textHandler("outside"))

	inside, err := statusOf(router, http.MethodGet, "/inside")
	if err != nil {
		return "", err
	}
	outside, err := statusOf(router, http.MethodGet, "/outside")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("routes=%d inside=%d outside=%d", len(router.List()), inside, outside), nil
}

// routeScopeBindingsScenario keeps the Route scoped-binding chain contracts at compile time.
func routeScopeBindingsScenario() (string, error) {
	var (
		_ func(*route.Route) *route.Route = (*route.Route).ScopeBindings
		_ func(*route.Route) *route.Route = (*route.Route).WithoutScopedBindings
	)
	router := route.New()
	configured := router.Get("/scoped", textHandler("scoped")).ScopeBindings().WithoutScopedBindings()
	return fmt.Sprintf("chain=%t routes=%d", configured != nil, len(router.List())), nil
}

// registrarScopeBindingsScenario keeps the Registrar scoped-binding chain contracts at compile time.
func registrarScopeBindingsScenario() (string, error) {
	var (
		_ func(*route.Registrar) *route.Registrar = (*route.Registrar).ScopeBindings
		_ func(*route.Registrar) *route.Registrar = (*route.Registrar).WithoutScopedBindings
	)
	router := route.New()
	registrar := router.Prefix("/api").ScopeBindings().WithoutScopedBindings()
	registrar.Get("/scoped", textHandler("scoped"))
	return fmt.Sprintf("chain=%t routes=%d", registrar != nil, len(router.List())), nil
}

// labelController lets the registrar override scenario observe which controller wins.
type labelController struct {
	label string
}

// Index writes the controller label selected by the innermost registrar.
func (l labelController) Index(c *gin.Context) { c.String(http.StatusOK, l.label) }

// registrarOverridesScenario shows that an inner Registrar Domain and Controller win.
func registrarOverridesScenario() (string, error) {
	router := route.New()
	router.Domain("outer.test").Controller(labelController{label: "outer"}).Group(func() {
		router.Domain("inner.test").Controller(labelController{label: "inner"}).
			Action(http.MethodGet, "/override", "Index")
	})

	domain := ""
	for _, info := range router.List() {
		domain = info.Domain
	}
	inner, err := dispatchHost(router, http.MethodGet, "/override", "inner.test")
	if err != nil {
		return "", err
	}
	outer, err := dispatchHost(router, http.MethodGet, "/override", "outer.test")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("domain=%s action=%s outer=%d", domain, inner.Body.String(), outer.Code), nil
}

// facadeContractScenario keeps every global facade function signature referenced at compile time.
func facadeContractScenario() (string, error) {
	var (
		_ func() *route.Router                                                              = route.Resolve
		_ func(string, route.Binder)                                                        = route.Bind
		_ func(string, route.Binder)                                                        = route.Model
		_ func(string, string)                                                              = route.Pattern
		_ func(*gin.Engine) error                                                           = route.Mount
		_ func() []route.RouteInfo                                                          = route.List
		_ func(string, map[string]any) (string, error)                                      = route.URL
		_ func(string, ...route.HandlerFunc) *route.Route                                   = route.Get
		_ func(string, ...route.HandlerFunc) *route.Route                                   = route.Post
		_ func(string, ...route.HandlerFunc) *route.Route                                   = route.Put
		_ func(string, ...route.HandlerFunc) *route.Route                                   = route.Patch
		_ func(string, ...route.HandlerFunc) *route.Route                                   = route.Delete
		_ func(string, ...route.HandlerFunc) *route.Route                                   = route.Options
		_ func([]string, string, ...route.HandlerFunc) *route.Route                         = route.Match
		_ func(string, ...route.HandlerFunc) *route.Route                                   = route.Any
		_ func(string, string, ...int) *route.Route                                         = route.Redirect
		_ func(string, string) *route.Route                                                 = route.PermanentRedirect
		_ func(string, string) *route.Route                                                 = route.Static
		_ func(route.HandlerFunc) *route.Route                                              = route.Fallback
		_ func(string) *route.Registrar                                                     = route.Prefix
		_ func(string) *route.Registrar                                                     = route.Name
		_ func(string) *route.Registrar                                                     = route.Domain
		_ func(...route.HandlerFunc) *route.Registrar                                       = route.Middleware
		_ func(...string) *route.Registrar                                                  = route.WithoutMiddleware
		_ func(any) *route.Registrar                                                        = route.Controller
		_ func(func())                                                                      = route.Group
		_ func(string, route.ResourceController, ...route.ResourceOption) []*route.Route    = route.Resource
		_ func(string, route.ResourceController, ...route.ResourceOption) []*route.Route    = route.ApiResource
		_ func(map[string]route.ResourceController, ...route.ResourceOption) []*route.Route = route.ApiResources
	)
	return "facade=Resolve,Bind,Model,Pattern,Mount,List,URL,Get,Post,Put,Patch,Delete,Options,Match,Any," +
		"Redirect,PermanentRedirect,Static,Fallback,Prefix,Name,Domain,Middleware,WithoutMiddleware," +
		"Controller,Group,Resource,ApiResource,ApiResources", nil
}
