package exceptiondemo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/exception"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
)

type requestOptions struct {
	useMiddleware bool
	requestID     string
	register      func(*gin.Engine)
}

type httpResult struct {
	status int
	body   string
	header http.Header
}

func serve(handler *exception.Handler, options requestOptions) (httpResult, error) {
	if options.register == nil {
		return httpResult{}, fmt.Errorf("register exception demo route: nil registrar")
	}
	// These in-process examples run as one CLI command; release mode keeps Gin's
	// route diagnostics off stdout so --json remains a single JSON document.
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if options.requestID != "" {
		engine.Use(httpmiddleware.RequestID())
	}
	if options.useMiddleware {
		engine.Use(httpmiddleware.Exception(handler))
	}
	options.register(engine)
	request := httptest.NewRequest(http.MethodGet, "/failure", nil)
	if options.requestID != "" {
		request.Header.Set("X-Request-ID", options.requestID)
	}
	writer := httptest.NewRecorder()
	engine.ServeHTTP(writer, request)
	return httpResult{status: writer.Code, body: writer.Body.String(), header: writer.Header()}, nil
}

func render(handler *exception.Handler, err error, requestID string) (httpResult, error) {
	return serve(handler, requestOptions{requestID: requestID, register: func(engine *gin.Engine) {
		engine.GET("/failure", func(c *gin.Context) { handler.Render(c, err) })
	}})
}

func problem(response httpResult) (exception.Problem, error) {
	var result exception.Problem
	if err := json.Unmarshal([]byte(response.body), &result); err != nil {
		return result, fmt.Errorf("decode HTTP %d Problem response: %w", response.status, err)
	}
	return result, nil
}

func middlewareConfiguration() (string, error) {
	h := exception.New(exception.WithLogging(false))
	register := func(engine *gin.Engine) {
		engine.GET("/failure", func(c *gin.Context) { _ = c.Error(errDemo) })
	}
	on, err := serve(h, requestOptions{useMiddleware: true, register: register})
	if err != nil {
		return "", err
	}
	off, err := serve(h, requestOptions{register: register})
	if err != nil {
		return "", err
	}
	direct, err := render(h, errDemo, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("configured=%t; enabled=%d; disabled=%d; direct=%d", config.GetBool("app.server.exception_handler"), on.status, off.status, direct.status), nil
}

func renderCase(name string) (string, error) {
	noLog := exception.WithLogging(false)
	switch name {
	case "default-render":
		response, err := render(exception.New(noLog), errDemo, "")
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("status=%d; type=%s; detail=%s", response.status, p.Type, p.Detail), err
	case "renderer-chain":
		order := []string{}
		h := exception.New(noLog,
			exception.WithRenderer(func(*gin.Context, error) (exception.Problem, bool) {
				order = append(order, "skip")
				return exception.Problem{}, false
			}),
			exception.WithRenderer(func(*gin.Context, error) (exception.Problem, bool) {
				order = append(order, "handle")
				return exception.Problem{Type: "custom", Title: "Conflict", Status: 409}, true
			}),
		)
		response, err := render(h, errDemo, "")
		if err != nil {
			return "", err
		}
		fallback, err := render(exception.New(noLog, exception.WithRenderer(func(*gin.Context, error) (exception.Problem, bool) {
			return exception.Problem{}, false
		})), errDemo, "")
		return fmt.Sprintf("order=%s; status=%d; fallback=%d", strings.Join(order, ","), response.status, fallback.status), err
	case "problem-renderer", "problem-response":
		h := exception.New(noLog, exception.WithRenderer(func(_ *gin.Context, err error) (exception.Problem, bool) {
			if _, ok := err.(demoHTTPError); !ok {
				return exception.Problem{}, false
			}
			return exception.Problem{Type: "order_conflict", Title: "Conflict", Status: 409, Code: 1001, Message: "order already closed"}, true
		}))
		response, err := render(h, demoHTTPError{status: 409, message: "order already closed"}, "")
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("status=%d; type=%s; code=%d; message=%s", response.status, p.Type, p.Code, p.Message), err
	case "response-renderer":
		h := exception.New(noLog, exception.WithResponseRenderer(func(c *gin.Context, _ error) bool {
			c.Data(503, "text/html; charset=utf-8", []byte("<h1>Service unavailable</h1>"))
			return true
		}))
		response, err := render(h, errDemo, "")
		return fmt.Sprintf("status=%d; html=%t", response.status, strings.Contains(response.body, "<h1>")), err
	case "response-fallback":
		h := exception.New(noLog, exception.WithResponseRenderer(func(*gin.Context, error) bool { return false }))
		response, err := render(h, errDemo, "")
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("status=%d; type=%s", response.status, p.Type), err
	case "panic-recovery":
		panicked := false
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, _ error, fields map[string]any) {
			panicked = fields["panic"] == "secret panic"
		}))
		response, err := serve(h, requestOptions{useMiddleware: true, register: func(engine *gin.Engine) {
			engine.GET("/failure", func(*gin.Context) { panic("secret panic") })
		}})
		return fmt.Sprintf("status=%d; reported=%t; safe=%t", response.status, panicked, !strings.Contains(response.body, "secret panic")), err
	case "gin-errors":
		var errorsField any
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, _ error, fields map[string]any) { errorsField = fields["errors"] }))
		response, err := serve(h, requestOptions{useMiddleware: true, register: func(engine *gin.Engine) {
			engine.GET("/failure", func(c *gin.Context) {
				_ = c.Error(errDemo)
				_ = c.Error(demoHTTPError{status: 409, message: "conflict"})
			})
		}})
		return fmt.Sprintf("status=%d; joined=%t", response.status, strings.Contains(fmt.Sprint(errorsField), " | ")), err
	case "status-report":
		statuses := []int{}
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, _ error, fields map[string]any) { statuses = append(statuses, fields["status"].(int)) }))
		for _, status := range []int{404, 503} {
			_, err := serve(h, requestOptions{useMiddleware: true, register: func(engine *gin.Engine) {
				engine.GET("/failure", func(c *gin.Context) { c.Status(status) })
			}})
			if err != nil {
				return "", err
			}
		}
		return fmt.Sprintf("reported=%v", statuses), nil
	case "request-id":
		var logged any
		h := exception.New(exception.WithPanicStack(false), exception.WithReporter(func(_ any, _ error, fields map[string]any) { logged = fields["request_id"] }))
		response, err := serve(h, requestOptions{useMiddleware: true, requestID: "demo-request-42", register: func(engine *gin.Engine) {
			engine.GET("/failure", func(c *gin.Context) { _ = c.Error(errDemo) })
		}})
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("header=%s; problem=%s; log=%v", response.header.Get("X-Request-ID"), p.RequestID, logged), err
	case "problem-optional":
		validation := demoHTTPError{status: 422, message: "validation failed", fields: map[string]any{"email": "required"}}
		without, err := render(exception.New(noLog), validation, "")
		if err != nil {
			return "", err
		}
		with, err := render(exception.New(noLog), validation, "demo-request-42")
		if err != nil {
			return "", err
		}
		first, err := problem(without)
		if err != nil {
			return "", err
		}
		second, err := problem(with)
		return fmt.Sprintf("error=%v; omitted=%t; request_id=%s", first.Errors["email"], first.RequestID == "", second.RequestID), err
	case "problem-fields":
		plain, err := render(exception.New(noLog), errDemo, "")
		if err != nil {
			return "", err
		}
		debug, err := render(exception.New(noLog, exception.WithDebug(true)), errDemo, "")
		if err != nil {
			return "", err
		}
		var public, diagnostic map[string]any
		if err := json.Unmarshal([]byte(plain.body), &public); err != nil {
			return "", err
		}
		if err := json.Unmarshal([]byte(debug.body), &diagnostic); err != nil {
			return "", err
		}
		return fmt.Sprintf("public_status=%v; plain_trace=%t; debug_trace=%t", public["status"], public["trace"] != nil, diagnostic["trace"] != nil), nil
	case "http-error":
		response, err := render(exception.New(noLog), demoHTTPError{status: 404, message: "order missing"}, "")
		if err != nil {
			return "", err
		}
		p, err := problem(response)
		return fmt.Sprintf("status=%d; detail=%s", response.status, p.Detail), err
	case "public-detail":
		client, err := render(exception.New(noLog), demoHTTPError{status: 422, message: "invalid order"}, "")
		if err != nil {
			return "", err
		}
		server, err := render(exception.New(noLog, exception.WithDebug(true)), demoHTTPError{status: 503, message: "database password"}, "")
		if err != nil {
			return "", err
		}
		debug, err := render(exception.New(noLog, exception.WithDebug(true)), errDemo, "")
		if err != nil {
			return "", err
		}
		first, err := problem(client)
		if err != nil {
			return "", err
		}
		second, err := problem(server)
		return fmt.Sprintf("client=%s; server=%s; debug_internal=%t", first.Detail, second.Detail, strings.Contains(debug.body, errDemo.Error())), err
	default:
		return "", fmt.Errorf("unknown rendering scenario %q", name)
	}
}
