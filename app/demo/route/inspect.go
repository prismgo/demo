package routedemo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/cmd"
	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/route"
)

// currentRouteScenario reads the injected RouteInfo inside the business handler.
func currentRouteScenario() (string, error) {
	router := route.New()
	router.Get("/users/{id}", func(c *gin.Context) {
		value, ok := c.Get("route.current")
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		info, ok := value.(route.RouteInfo)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("%s|%s|%s", info.Name, info.URI, info.GinPath))
	}).Name("users.show")

	body, err := bodyOf(router, http.MethodGet, "/users/7")
	if err != nil {
		return "", err
	}
	return "info=" + body, nil
}

// routeInfoScenario keeps the documented RouteInfo field contract referenced at compile time.
func routeInfoScenario() (string, error) {
	infoType := reflect.TypeOf(route.RouteInfo{})
	fields := make([]string, 0, infoType.NumField())
	for index := 0; index < infoType.NumField(); index++ {
		field := infoType.Field(index)
		fields = append(fields, field.Name+":"+field.Type.String())
	}
	return fmt.Sprintf("interface=%s fields=%s", infoType.Name(), strings.Join(fields, ",")), nil
}

// listScenario snapshots every declared route through Router.List.
func listScenario() (string, error) {
	router := route.New()
	router.Get("/users", textHandler("index")).Name("users.index")
	router.Get("/users/{id}", textHandler("show")).Name("users.show")

	infos := router.List()
	return fmt.Sprintf("routes=%d %s", len(infos), routeListSummary(infos)), nil
}

// listCommandScenario drives the route:list command with a name filter and JSON output.
func listCommandScenario() (string, error) {
	router := route.Resolve()
	router.Reset()
	defer router.Reset()

	route.Prefix("/api/v1").Name("api.").Group(func() {
		route.Get("/users/{id}", textHandler("show")).Name("users.show")
		route.Post("/users", textHandler("store")).Name("users.store")
	})

	command := cmd.NewRouteListCommand(func() error { return nil })
	var output bytes.Buffer
	input := scenarioInput{options: map[string]string{"name": "show"}, bools: map[string]bool{"json": true}}
	ctx := console.NewCommandContext(context.Background(), command, *command.Definition(), input,
		console.NewIO(strings.NewReader(""), &output, io.Discard), nil, nil)
	if err := command.Handle(ctx); err != nil {
		return "", fmt.Errorf("run route:list: %w", err)
	}

	var rows []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(output.Bytes(), &rows); err != nil {
		return "", fmt.Errorf("decode route:list json: %w", err)
	}
	if len(rows) != 1 {
		return "", fmt.Errorf("route:list filtered rows = %d, want 1", len(rows))
	}
	return fmt.Sprintf("routes=%d name=%s", len(rows), rows[0].Name), nil
}

// handlerOrderScenario records the group, route middleware and action ordering.
func handlerOrderScenario() (string, error) {
	router := route.New()
	trace := make([]string, 0, 3)
	group := route.NamedMiddleware("demo-order-group", func(c *gin.Context) {
		trace = append(trace, "group")
		c.Next()
	})
	stage := func(c *gin.Context) {
		trace = append(trace, "route")
		c.Next()
	}
	router.Middleware(group).Group(func() {
		router.Get("/ordered", stage, func(c *gin.Context) {
			trace = append(trace, "action")
			c.String(http.StatusOK, "ok")
		})
	})

	if _, err := dispatch(router, http.MethodGet, "/ordered"); err != nil {
		return "", err
	}
	return "order=" + strings.Join(trace, ">"), nil
}

// routeListSummary renders the method, URI and name of each route snapshot.
func routeListSummary(infos []route.RouteInfo) string {
	parts := make([]string, 0, len(infos))
	for _, info := range infos {
		parts = append(parts, fmt.Sprintf("%s %s=%s", strings.Join(info.Methods, ","), displayRouteURI(info), info.Name))
	}
	return strings.Join(parts, ";")
}

// displayRouteURI prefers the declared URI and falls back to the compiled Gin path.
func displayRouteURI(info route.RouteInfo) string {
	if info.URI != "" {
		return info.URI
	}
	return info.GinPath
}

// scenarioInput is a minimal console.Input that drives commands from route scenarios.
type scenarioInput struct {
	arguments map[string]string
	options   map[string]string
	bools     map[string]bool
}

// Argument returns a scalar positional argument.
func (i scenarioInput) Argument(name string) string { return i.arguments[name] }

// Arguments returns no variadic arguments.
func (i scenarioInput) Arguments(string) []string { return nil }

// Option returns a scalar option value.
func (i scenarioInput) Option(name string) string { return i.options[name] }

// OptionStrings returns no repeated option values.
func (i scenarioInput) OptionStrings(string) []string { return nil }

// OptionBool returns a boolean option value.
func (i scenarioInput) OptionBool(name string) bool { return i.bools[name] }

// OptionInt returns the zero value for unused integer options.
func (i scenarioInput) OptionInt(string) (int, error) { return 0, nil }

// HasOption reports whether a scalar option is defined.
func (i scenarioInput) HasOption(name string) bool { _, ok := i.options[name]; return ok }
