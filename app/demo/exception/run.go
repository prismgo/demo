// Package exceptiondemo contains runnable exception reporting and HTTP rendering examples.
package exceptiondemo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/exception"
	"github.com/prismgo/framework/foundation"
)

// Result records an observable exception scenario outcome.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

var errDemo = errors.New("demo internal secret")

type demoHTTPError struct {
	status  int
	message string
	fields  map[string]any
}

func (e demoHTTPError) Error() string                { return "internal cause: " + e.message }
func (e demoHTTPError) StatusCode() int              { return e.status }
func (e demoHTTPError) PublicMessage() string        { return e.message }
func (e demoHTTPError) PublicFields() map[string]any { return e.fields }

// Run executes one documented exception scenario in the current application.
func Run(name string) (Result, error) {
	var value string
	var err error
	switch name {
	case "architecture":
		h := exception.Resolve()
		if h == nil {
			return Result{}, fmt.Errorf("exception handler is not registered")
		}
		response, renderErr := render(h, errDemo, "")
		value, err = fmt.Sprintf("handler=%T; should_report=%t; render=%d", h, h.ShouldReport(errDemo, 500), response.status), renderErr
	case "config":
		value = fmt.Sprintf("app.debug=%t; exception_handler=%t", config.GetBool("app.debug"), config.GetBool("app.server.exception_handler"))
	case "debug-config":
		value = fmt.Sprintf("configured=%t; standalone=%t", exception.Resolve().Debug(), exception.New().Debug())
	case "middleware-config":
		value, err = middlewareConfiguration()
	case "default-report":
		value, err = defaultReport()
	case "custom-reporter":
		value, err = customReporter()
	case "reporter-order":
		value, err = reporterOrder()
	case "package-report":
		value, err = packageReport()
	case "log-context":
		value, err = logContext()
	case "default-level":
		h := exception.New()
		value = fmt.Sprintf("500=%s; 422=%s", h.Level(errDemo, 500), h.Level(errDemo, 422))
	case "custom-level":
		h := exception.New(exception.WithLevel(func(error, int) exception.Level { return exception.LevelInfo }))
		custom := h.Level(errDemo, 500)
		h.ApplyOptions(exception.WithLevel(func(error, int) exception.Level { return "" }))
		value = fmt.Sprintf("custom=%s; fallback=%s", custom, h.Level(errDemo, 500))
	case "level-constants":
		value = strings.Join([]string{string(exception.LevelDebug), string(exception.LevelInfo), string(exception.LevelWarn), string(exception.LevelError)}, ",")
	case "dont-report":
		value = ignoredReport()
	case "predicate-order":
		value = predicateOrder()
	case "context-ignored":
		h := exception.New()
		value = fmt.Sprintf("canceled=%t; deadline=%t", h.ShouldReport(context.Canceled, 500), h.ShouldReport(context.DeadlineExceeded, 500))
	case "clear-ignored":
		h := exception.New()
		h.DontReport = nil
		value = fmt.Sprintf("canceled=%t", h.ShouldReport(context.Canceled, 500))
	case "default-render", "renderer-chain", "problem-renderer", "response-renderer", "response-fallback", "panic-recovery", "gin-errors", "status-report", "request-id", "problem-response", "problem-optional", "problem-fields", "http-error", "public-detail":
		value, err = renderCase(name)
	default:
		return Result{}, fmt.Errorf("unknown exception scenario %q", name)
	}
	if err != nil {
		return Result{}, fmt.Errorf("exception scenario %q: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func customReporter() (value string, err error) {
	var observed string
	app := foundation.Configure().WithExceptions(func(e *foundation.Exceptions) {
		e.Report(func(ctx any, err error, fields map[string]any) {
			observed = fmt.Sprintf("context=%T; error=%t; status=%v", ctx, errors.Is(err, errDemo), fields["status"])
		})
	}).Create()
	defer func() {
		if closeErr := app.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close custom reporter application: %w", closeErr))
		}
	}()
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot custom reporter application: %w", err)
	}
	raw, err := app.Container().Make(foundation.ContainerKeyExceptionHandler)
	if err != nil {
		return "", fmt.Errorf("resolve custom reporter handler: %w", err)
	}
	h, ok := raw.(*exception.Handler)
	if !ok {
		return "", fmt.Errorf("custom reporter handler = %T, want *exception.Handler", raw)
	}
	h.Report(context.Background(), errDemo, map[string]any{"status": 503})
	return observed, nil
}

func ignoredReport() string {
	calls := 0
	h := exception.New(exception.WithDontReport(func(err error) bool { return errors.Is(err, errDemo) }), exception.WithReporter(func(any, error, map[string]any) { calls++ }))
	h.Report(context.Background(), errDemo, nil)
	return fmt.Sprintf("should_report=%t; reporters=%d", h.ShouldReport(errDemo, http.StatusInternalServerError), calls)
}

func predicateOrder() string {
	order := []string{}
	h := exception.New(
		exception.WithDontReport(func(error) bool { order = append(order, "first"); return false }),
		exception.WithDontReport(func(error) bool { order = append(order, "second"); return true }),
		exception.WithDontReport(func(error) bool { order = append(order, "third"); return true }),
	)
	_ = h.ShouldReport(errDemo, 500)
	return strings.Join(order, ",")
}
