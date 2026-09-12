package exceptiondemo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/exception"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
)

type demoBusinessError struct{}

func (demoBusinessError) Error() string                { return "internal business cause" }
func (demoBusinessError) StatusCode() int              { return http.StatusConflict }
func (demoBusinessError) PublicMessage() string        { return "order conflict" }
func (demoBusinessError) PublicFields() map[string]any { return map[string]any{"order": "closed"} }
func (demoBusinessError) BusinessCode() int            { return 1001 }
func (demoBusinessError) ErrorType() string            { return "order_conflict" }
func (demoBusinessError) ErrorContext() map[string]any { return map[string]any{"order_id": 42} }

func secondBatch(name string) (string, error) {
	switch name {
	case "safe-response":
		response, err := render(exception.New(exception.WithLogging(false)), fmt.Errorf("password=secret; token=private: %w", errDemo), "")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("status=%d; concealed=%t", response.status, !strings.Contains(response.body, "secret") && !strings.Contains(response.body, "password") && !strings.Contains(response.body, "token") && !strings.Contains(response.body, "demo internal")), nil
	case "request-fields":
		var fields map[string]any
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, _ error, reported map[string]any) { fields = reported }))
		response, err := serve(h, requestOptions{useMiddleware: true, target: "/failure?source=demo", register: errorRoute(errDemo)})
		if err != nil {
			return "", err
		}
		_, durationOK := fields["duration_ms"].(int64)
		return fmt.Sprintf("status=%d; method=%v; path=%v; url=%v; query=%v; ip=%t; duration=%t", response.status, fields["method"], fields["path"], fields["url"], fields["query"], fields["client_ip"] != "", durationOK), nil
	case "diagnostic-fields":
		return diagnosticFields()
	case "business-fields":
		var fields map[string]any
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, _ error, reported map[string]any) { fields = reported }))
		response, err := serve(h, requestOptions{useMiddleware: true, register: errorRoute(demoBusinessError{})})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("status=%d; code=%v; type=%v; context=%v; fields=%v", response.status, fields["error_code"], fields["error_type"], fields["error_context"], fields["field_errors"]), nil
	case "handler-defaults":
		h := exception.New()
		return fmt.Sprintf("ignore=%d; recovery=%t; logging=%t; client=%t; stack=%t; debug=%t", len(h.DontReport), h.RecoverPanics, h.LogErrors, h.LogClientErrors, h.PanicStack, h.Debug()), nil
	case "handler-flags":
		h := exception.New(exception.WithLogging(false), exception.WithClientErrorLogging(false), exception.WithRecovery(false), exception.WithPanicStack(false))
		return fmt.Sprintf("report=%t; client=%t; recovery=%t; stack=%t", h.ShouldReport(errDemo, 500), h.ShouldReport(errDemo, 422), h.RecoverPanics, h.PanicStack), nil
	case "options":
		h := exception.New(exception.WithLogging(false))
		h.ApplyOptions(exception.WithLogging(true), exception.WithDebug(true))
		return fmt.Sprintf("logging=%t; debug=%t", h.LogErrors, h.Debug()), nil
	case "with-dont-report", "predicate-type":
		var predicate exception.Predicate = func(err error) bool { return errors.Is(err, errDemo) }
		h := exception.New(exception.WithDontReport(predicate))
		return fmt.Sprintf("matched=%t; other=%t", !h.ShouldReport(errDemo, 500), h.ShouldReport(errors.New("other"), 500)), nil
	case "with-level", "level-resolver-type":
		var resolver exception.LevelResolver = func(_ error, status int) exception.Level {
			if status == 503 {
				return exception.LevelInfo
			}
			return ""
		}
		h := exception.New(exception.WithLevel(resolver))
		return fmt.Sprintf("custom=%s; fallback=%s", h.Level(errDemo, 503), h.Level(errDemo, 500)), nil
	case "with-reporter":
		calls := 0
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(any, error, map[string]any) { calls++ }))
		h.Report(context.Background(), errDemo, map[string]any{"status": 500})
		return fmt.Sprintf("calls=%d", calls), nil
	case "with-recovery":
		return recoveryOption()
	case "with-logging":
		h := exception.New(exception.WithLogging(false), exception.WithReporter(func(any, error, map[string]any) {}))
		return fmt.Sprintf("should_report=%t", h.ShouldReport(errDemo, 500)), nil
	case "with-client-logging":
		h := exception.New(exception.WithClientErrorLogging(false))
		return fmt.Sprintf("client=%t; server=%t", h.ShouldReport(errDemo, 422), h.ShouldReport(errDemo, 500)), nil
	case "with-panic-stack":
		return panicStackOption()
	case "with-debug":
		h := exception.New(exception.WithDebug(true), exception.WithLogging(false))
		response, err := render(h, errDemo, "")
		return fmt.Sprintf("debug=%t; exposed=%t", h.Debug(), strings.Contains(response.body, errDemo.Error())), err
	case "with-debug-resolver":
		enabled := false
		h := exception.New(exception.WithDebugResolver(func() bool { return enabled }))
		before := h.Debug()
		enabled = true
		return fmt.Sprintf("before=%t; after=%t", before, h.Debug()), nil
	case "with-context", "context-extractor-type":
		return extractedContext()
	case "with-renderer", "renderer-type":
		var renderer exception.Renderer = func(*gin.Context, error) (exception.Problem, bool) {
			return exception.Problem{Type: "demo", Title: "Conflict", Status: 409}, true
		}
		response, err := render(exception.New(exception.WithLogging(false), exception.WithRenderer(renderer)), errDemo, "")
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("status=%d; type=%s", response.status, p.Type), err
	case "with-response-renderer", "response-renderer-type":
		var renderer exception.ResponseRenderer = func(c *gin.Context, _ error) bool {
			c.Data(503, "text/plain", []byte("maintenance"))
			return true
		}
		response, err := render(exception.New(exception.WithLogging(false), exception.WithResponseRenderer(renderer)), errDemo, "")
		return fmt.Sprintf("status=%d; body=%s", response.status, response.body), err
	case "reporter-type":
		return reporterContract()
	case "resolve":
		h := exception.Resolve()
		return fmt.Sprintf("registered=%t; type=%T", h != nil, h), nil
	case "facade-report":
		return facadeReport()
	case "facade-render":
		response, err := serve(exception.New(exception.WithLogging(false)), requestOptions{register: func(engine *gin.Engine) {
			engine.GET("/failure", func(c *gin.Context) { exception.Render(c, errDemo) })
		}})
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("status=%d; type=%s", response.status, p.Type), err
	case "build-register":
		called := false
		h := exception.BuildAndRegister([]exception.Option{exception.WithLogging(false)}, func(h *exception.Handler) *exception.Handler {
			called = true
			h.ApplyOptions(exception.WithDebug(true))
			return h
		})
		return fmt.Sprintf("factory=%t; logging=%t; debug=%t", called, h.LogErrors, h.Debug()), nil
	case "handler-report":
		var observed error
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, err error, _ map[string]any) { observed = err }))
		h.Report(context.Background(), errDemo, map[string]any{"status": 500})
		return fmt.Sprintf("reported=%t", errors.Is(observed, errDemo)), nil
	default:
		return "", fmt.Errorf("unknown second-batch exception scenario %q", name)
	}
}

func errorRoute(err error) func(*gin.Engine) {
	return func(engine *gin.Engine) {
		engine.GET("/failure", func(c *gin.Context) { _ = c.Error(err) })
	}
}

func diagnosticFields() (string, error) {
	var errorFields, panicFields map[string]any
	h := exception.New(exception.WithPanicStack(true), exception.WithReporter(func(_ any, _ error, fields map[string]any) {
		if fields["panic"] != nil {
			panicFields = fields
			return
		}
		errorFields = fields
	}))
	_, err := serve(h, requestOptions{useMiddleware: true, requestID: "diagnostic-42", register: errorRoute(errDemo)})
	if err != nil {
		return "", err
	}
	_, err = serve(h, requestOptions{useMiddleware: true, register: func(engine *gin.Engine) {
		engine.GET("/failure", func(*gin.Context) { panic("demo panic") })
	}})
	if err != nil {
		return "", err
	}
	logContent, err := readErrorLog()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("request_id=%v; errors=%t; panic=%v; stack=%t", errorFields["request_id"], strings.Contains(fmt.Sprint(errorFields["errors"]), errDemo.Error()), panicFields["panic"], strings.Contains(logContent, "panic recovered: demo panic") && strings.Contains(logContent, "[stacktrace]")), nil
}

func recoveryOption() (string, error) {
	h := exception.New(exception.WithRecovery(false), exception.WithLogging(false))
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	var propagated any
	engine.Use(func(c *gin.Context) {
		defer func() { propagated = recover() }()
		c.Next()
	})
	// The outer catcher observes the panic propagated by Exception.
	engine.Use(httpmiddleware.Exception(h))
	engine.GET("/failure", func(*gin.Context) { panic("propagated") })
	serveEngine(engine)
	return fmt.Sprintf("recovery=%t; propagated=%v", h.RecoverPanics, propagated), nil
}

func panicStackOption() (string, error) {
	marker := fmt.Sprintf("exception stack option %d", time.Now().UnixNano())
	withoutCapture := exception.New(exception.WithPanicStack(false))
	withCapture := exception.New(exception.WithPanicStack(true))
	withoutCapture.Report(context.Background(), errors.New(marker+" disabled"), map[string]any{"status": 500})
	withCapture.Report(context.Background(), errors.New(marker+" enabled"), map[string]any{"status": 500})
	content, err := readErrorLog()
	if err != nil {
		return "", err
	}
	first := strings.LastIndex(content, marker+" disabled")
	second := strings.LastIndex(content, marker+" enabled")
	if first < 0 || second <= first {
		return "", fmt.Errorf("stack option log entries not found in order: disabled=%d, enabled=%d", first, second)
	}
	// The line formatter can still capture a stack when Handler capture is off.
	return fmt.Sprintf("handler_capture=%t,%t; log_stack=%t,%t", withoutCapture.PanicStack, withCapture.PanicStack, strings.Contains(content[first:second], "[stacktrace]"), strings.Contains(content[second:], "[stacktrace]")), nil
}

func extractedContext() (string, error) {
	var fields map[string]any
	var extractor exception.ContextExtractor = func(*gin.Context) map[string]any {
		return map[string]any{"tenant_id": "tenant-42"}
	}
	h := exception.New(exception.WithPanicStack(false), exception.WithContext(extractor), exception.WithReporter(func(_ any, _ error, reported map[string]any) { fields = reported }))
	response, err := serve(h, requestOptions{useMiddleware: true, register: errorRoute(errDemo)})
	return fmt.Sprintf("status=%d; tenant=%v; response_private=%t", response.status, fields["tenant_id"], !strings.Contains(response.body, "tenant_id")), err
}

func reporterContract() (string, error) {
	var contexts []string
	var statuses []any
	sameError := true
	var reporter exception.Reporter = func(ctx any, err error, fields map[string]any) {
		contexts = append(contexts, fmt.Sprintf("%T", ctx))
		sameError = sameError && errors.Is(err, errDemo)
		statuses = append(statuses, fields["status"])
	}
	h := exception.New(exception.WithPanicStack(false), exception.WithReporter(reporter))
	h.Report(context.Background(), errDemo, map[string]any{"status": 503})
	_, err := serve(h, requestOptions{useMiddleware: true, register: errorRoute(errDemo)})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("contexts=%s; error=%t; statuses=%v", strings.Join(contexts, ","), sameError, statuses), nil
}

func facadeReport() (string, error) {
	h := exception.Resolve()
	if h == nil {
		return "", fmt.Errorf("resolve exception facade handler: nil handler")
	}
	called := false
	h.ApplyOptions(exception.WithReporter(func(_ any, err error, _ map[string]any) { called = errors.Is(err, errDemo) }))
	exception.Report(context.Background(), errDemo, map[string]any{"status": 500})
	return fmt.Sprintf("reported=%t", called), nil
}

func serveEngine(engine *gin.Engine) httpResult {
	request := httptest.NewRequest(http.MethodGet, "/failure", nil)
	writer := httptest.NewRecorder()
	engine.ServeHTTP(writer, request)
	return httpResult{status: writer.Code, body: writer.Body.String(), header: writer.Header()}
}
