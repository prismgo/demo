package routedemo

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// namedMiddlewareScenario excludes a named group middleware on a single route.
func namedMiddlewareScenario() (string, error) {
	router := route.New()
	auth := route.NamedMiddleware("auth", func(c *gin.Context) {
		c.Header("X-Auth", "on")
		c.Next()
	})
	router.Middleware(auth).Group(func() {
		router.Get("/profile", textHandler("profile"))
		router.Get("/public", textHandler("public")).WithoutMiddleware("auth")
	})

	profile, err := dispatch(router, http.MethodGet, "/profile")
	if err != nil {
		return "", err
	}
	public, err := dispatch(router, http.MethodGet, "/public")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("profile=%s public=%s names=%s",
		headerFlag(profile.Header().Get("X-Auth")),
		headerFlag(public.Header().Get("X-Auth")),
		strings.Join(middlewareNames(router.List()), ",")), nil
}

// middlewareFunctionNameScenario excludes an unnamed middleware by its Go function name.
func middlewareFunctionNameScenario() (string, error) {
	router := route.New()
	name := functionName(middlewareNameGuard)
	router.Middleware(middlewareNameGuard).Group(func() {
		router.Get("/guarded", textHandler("guarded"))
		router.Get("/open", textHandler("open")).WithoutMiddleware(name)
	})

	guarded, err := dispatch(router, http.MethodGet, "/guarded")
	if err != nil {
		return "", err
	}
	open, err := dispatch(router, http.MethodGet, "/open")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("name=%s guarded=%s open=%s",
		shortFunctionName(name),
		headerFlag(guarded.Header().Get("X-Guard")),
		headerFlag(open.Header().Get("X-Guard"))), nil
}

// routeWithoutMiddlewareScenario removes a route-level middleware while keeping group middleware.
func routeWithoutMiddlewareScenario() (string, error) {
	router := route.New()
	group := route.NamedMiddleware("group", func(c *gin.Context) {
		c.Header("X-Group", "on")
		c.Next()
	})
	local := route.NamedMiddleware("local", func(c *gin.Context) {
		c.Header("X-Local", "on")
		c.Next()
	})
	router.Middleware(group).Group(func() {
		router.Get("/both", local, textHandler("both"))
		router.Get("/group-only", local, textHandler("group-only")).WithoutMiddleware("local")
	})

	both, err := dispatch(router, http.MethodGet, "/both")
	if err != nil {
		return "", err
	}
	groupOnly, err := dispatch(router, http.MethodGet, "/group-only")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("both=group:%s,local:%s group-only=group:%s,local:%s",
		headerFlag(both.Header().Get("X-Group")), headerFlag(both.Header().Get("X-Local")),
		headerFlag(groupOnly.Header().Get("X-Group")), headerFlag(groupOnly.Header().Get("X-Local"))), nil
}

// registrarWithoutMiddlewareScenario excludes a group middleware at declaration time.
func registrarWithoutMiddlewareScenario() (string, error) {
	router := route.New()
	auth := route.NamedMiddleware("auth", func(c *gin.Context) {
		c.Header("X-Auth", "on")
		c.Next()
	})
	audit := route.NamedMiddleware("audit", func(c *gin.Context) {
		c.Header("X-Audit", "on")
		c.Next()
	})
	router.Middleware(auth, audit).Group(func() {
		router.Get("/full", textHandler("full"))
		router.WithoutMiddleware("audit").Get("/callback", textHandler("callback"))
	})

	full, err := dispatch(router, http.MethodGet, "/full")
	if err != nil {
		return "", err
	}
	callback, err := dispatch(router, http.MethodGet, "/callback")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("full=auth:%s,audit:%s callback=auth:%s,audit:%s",
		headerFlag(full.Header().Get("X-Auth")), headerFlag(full.Header().Get("X-Audit")),
		headerFlag(callback.Header().Get("X-Auth")), headerFlag(callback.Header().Get("X-Audit"))), nil
}

// middlewareNameGuard is an unnamed middleware used to exercise function-name exclusion.
func middlewareNameGuard(c *gin.Context) {
	c.Header("X-Guard", "on")
	c.Next()
}

// headerFlag reports whether a middleware wrote its response header.
func headerFlag(value string) string {
	if value == "" {
		return "off"
	}
	return "on"
}

// middlewareNames lists the distinct middleware identifiers recorded in route snapshots.
func middlewareNames(infos []route.RouteInfo) []string {
	seen := make(map[string]bool)
	names := make([]string, 0)
	for _, info := range infos {
		for _, name := range info.Middleware {
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// functionName resolves the fully qualified Go symbol used as the unnamed middleware fallback.
func functionName(handler route.HandlerFunc) string {
	value := reflect.ValueOf(handler)
	if !value.IsValid() || value.Kind() != reflect.Func {
		return ""
	}
	fn := runtime.FuncForPC(value.Pointer())
	if fn == nil {
		return ""
	}
	return fn.Name()
}

// shortFunctionName trims the package qualifier for a compact assertion.
func shortFunctionName(name string) string {
	if index := strings.LastIndex(name, "."); index >= 0 {
		return name[index+1:]
	}
	return name
}
