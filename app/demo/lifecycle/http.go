package lifecycledemo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/console"
	"github.com/prismgo/framework/event"
	"github.com/prismgo/framework/foundation"
	frameworkhttp "github.com/prismgo/framework/http"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
)

// newHTTPServer builds and boots an Application and returns its HTTP server.
func newHTTPServer(
	base string,
	routes func(*foundation.Application, *gin.Engine) error,
	middleware func(*foundation.Middleware),
	commands ...console.CommandFactory,
) (*foundation.Application, *http.Server, error) {
	app, err := buildHTTPApplication(base, routes, middleware, commands...)
	if err != nil {
		return nil, nil, err
	}
	server, err := app.NewHTTPServer(context.Background(), "0")
	if err != nil {
		_ = app.Close()
		return nil, nil, fmt.Errorf("construct application HTTP server: %w", err)
	}
	return app, server, nil
}

// serveRequest dispatches one request through the server handler.
func serveRequest(server *http.Server, method, path string, headers http.Header) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	for name, values := range headers {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	return response
}

// httpPipelineScenario verifies request events and middleware ordering around a handler.
func httpPipelineScenario(base string) (string, error) {
	log := &eventLog{}
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/pipeline", func(c *gin.Context) {
			log.add("handler")
			c.String(http.StatusOK, "ok")
		})
		return nil
	}
	middleware := func(m *foundation.Middleware) {
		m.Use(func(engine *gin.Engine) {
			engine.Use(func(c *gin.Context) {
				log.add("middleware")
				c.Next()
			})
		})
	}
	app, server, err := newHTTPServer(base, routes, middleware)
	if err != nil {
		return "", err
	}
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	bus.Listen(event.EventRequestReceived, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("received")
		return nil
	}))
	bus.Listen(event.EventRequestHandled, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("handled")
		return nil
	}))

	response := serveRequest(server, http.MethodGet, "/pipeline", nil)
	return fmt.Sprintf("order=%s status=%d", strings.Join(log.values(), ","), response.Code), nil
}

// requestIDScenario verifies request ID propagation through context and response header.
func requestIDScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/request-id", func(c *gin.Context) {
			c.String(http.StatusOK, "id=%s", frameworkhttp.GetRequestID(c))
		})
		return nil
	}
	middleware := func(m *foundation.Middleware) {
		m.Prepend(func(engine *gin.Engine) {
			engine.Use(httpmiddleware.RequestID())
		})
	}
	app, server, err := newHTTPServer(base, routes, middleware)
	if err != nil {
		return "", err
	}
	defer app.Close()

	response := serveRequest(server, http.MethodGet, "/request-id", nil)
	header := response.Header().Get(frameworkhttp.RequestIDHeader)
	body := response.Body.String()
	return fmt.Sprintf("header-present=%t body-echoes=%t", header != "", header != "" && strings.Contains(body, header)), nil
}

// accessLogScenario verifies the built-in access log records incoming requests.
func accessLogScenario(base string) (string, error) {
	restore := setEnv(map[string]string{"SERVER_ACCESS_LOG": "true"})
	defer restore()

	var output bytes.Buffer
	previous := gin.DefaultWriter
	gin.DefaultWriter = &output
	defer func() { gin.DefaultWriter = previous }()

	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/access", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		return nil
	}
	app, server, err := newHTTPServer(base, routes, nil)
	if err != nil {
		return "", err
	}
	defer app.Close()

	serveRequest(server, http.MethodGet, "/access", nil)
	return fmt.Sprintf("logged=%t", strings.Contains(output.String(), "/access")), nil
}

// exceptionMiddlewareScenario verifies recovery, rendering and error reporting.
func exceptionMiddlewareScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/error", func(c *gin.Context) {
			_ = c.Error(errors.New("demo internal failure"))
		})
		return nil
	}
	app, server, err := newHTTPServer(base, routes, nil)
	if err != nil {
		return "", err
	}
	defer app.Close()

	response := serveRequest(server, http.MethodGet, "/error", nil)
	return fmt.Sprintf("status=%d rendered=%t", response.Code, response.Body.Len() > 0), nil
}

// businessMiddlewareScenario verifies application middleware runs before route dispatch.
func businessMiddlewareScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/business", func(c *gin.Context) {
			value, _ := c.Get("demo.business")
			c.String(http.StatusOK, "business=%v", value)
		})
		return nil
	}
	middleware := func(m *foundation.Middleware) {
		m.Use(func(engine *gin.Engine) {
			engine.Use(func(c *gin.Context) {
				c.Set("demo.business", "ran")
				c.Next()
			})
		})
	}
	app, server, err := newHTTPServer(base, routes, middleware)
	if err != nil {
		return "", err
	}
	defer app.Close()

	response := serveRequest(server, http.MethodGet, "/business", nil)
	return strings.TrimSpace(response.Body.String()), nil
}

// requestErrorLogScenario verifies unhandled request errors reach the failure channel.
func requestErrorLogScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/failure", func(c *gin.Context) {
			_ = c.Error(errors.New("demo unhandled failure"))
		})
		return nil
	}
	app, server, err := newHTTPServer(base, routes, nil)
	if err != nil {
		return "", err
	}
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	bus.Listen(event.EventRequestFailed, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if failed, ok := ev.(event.RequestFailed); ok {
			log.add(fmt.Sprintf("%d:%s", failed.Status, failed.Error))
		}
		return nil
	}))

	response := serveRequest(server, http.MethodGet, "/failure", nil)
	entries := log.values()
	return fmt.Sprintf("failed-event=%t status=%d", len(entries) == 1, response.Code), nil
}

// serverEventRun records the server lifecycle events observed during one serve cycle.
type serverEventRun struct {
	sequence []string
	starting event.ServerStarting
	started  event.ServerStarted
	stopping event.ServerStopping
	stopped  event.ServerStopped
}

// observeServerEvents serves one request cycle and captures the server lifecycle events.
func observeServerEvents(base string) (serverEventRun, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/server", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		return nil
	}
	app, server, err := newHTTPServer(base, routes, nil)
	if err != nil {
		return serverEventRun{}, err
	}
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return serverEventRun{}, err
	}
	dispatcher, ok := bus.(*event.Dispatcher)
	if !ok {
		return serverEventRun{}, fmt.Errorf("lifecycle demo: event dispatcher has type %T, want *event.Dispatcher", bus)
	}

	observed := serverEventRun{}
	started := make(chan struct{})
	bus.Listen(event.EventServerStarting, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		observed.sequence = append(observed.sequence, "starting")
		if value, ok := ev.(event.ServerStarting); ok {
			observed.starting = value
		}
		return nil
	}))
	bus.Listen(event.EventServerStarted, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		observed.sequence = append(observed.sequence, "started")
		if value, ok := ev.(event.ServerStarted); ok {
			observed.started = value
		}
		close(started)
		return nil
	}))
	bus.Listen(event.EventServerStopping, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		observed.sequence = append(observed.sequence, "stopping")
		if value, ok := ev.(event.ServerStopping); ok {
			observed.stopping = value
		}
		return nil
	}))
	bus.Listen(event.EventServerStopped, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		observed.sequence = append(observed.sequence, "stopped")
		if value, ok := ev.(event.ServerStopped); ok {
			observed.stopped = value
		}
		return nil
	}))

	// Bind an ephemeral loopback listener so the scenario never depends on the
	// configured port; server lifecycle events keep reporting the documented :0 address.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return serverEventRun{}, fmt.Errorf("listen server events socket: %w", err)
	}
	server.Addr = ":0"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- frameworkhttp.ListenAndServeGracefulContext(ctx, server, 2*time.Second,
			frameworkhttp.WithDispatcher(dispatcher), frameworkhttp.WithListener(listener))
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		select {
		case serveErr := <-done:
			return serverEventRun{}, fmt.Errorf("lifecycle demo: server returned before start: %w (sequence=%v)", serveErr, observed.sequence)
		default:
		}
		return serverEventRun{}, fmt.Errorf("lifecycle demo: server did not start within timeout (sequence=%v)", observed.sequence)
	}
	cancel()
	select {
	case serveErr := <-done:
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serverEventRun{}, fmt.Errorf("serve server lifecycle events: %w", serveErr)
		}
	case <-time.After(5 * time.Second):
		return serverEventRun{}, fmt.Errorf("lifecycle demo: server did not stop within timeout")
	}
	return observed, nil
}

// serverStartingEventScenario verifies the server.starting payload and timing.
func serverStartingEventScenario(base string) (string, error) {
	observed, err := observeServerEvents(base)
	if err != nil {
		return "", err
	}
	before := indexOf(observed.sequence, "starting") == 0 && indexOf(observed.sequence, "started") > 0
	return fmt.Sprintf("addr=%s pid-positive=%t before-started=%t",
		observed.starting.Addr, observed.starting.PID > 0, before), nil
}

// serverStartedEventScenario verifies the server.started payload and timing.
func serverStartedEventScenario(base string) (string, error) {
	observed, err := observeServerEvents(base)
	if err != nil {
		return "", err
	}
	after := indexOf(observed.sequence, "started") > indexOf(observed.sequence, "starting")
	return fmt.Sprintf("addr=%s pid-positive=%t after-starting=%t",
		observed.started.Addr, observed.started.PID > 0, after), nil
}

// serverStoppingEventScenario verifies the server.stopping payload and shutdown reason.
func serverStoppingEventScenario(base string) (string, error) {
	observed, err := observeServerEvents(base)
	if err != nil {
		return "", err
	}
	after := indexOf(observed.sequence, "stopping") > indexOf(observed.sequence, "started")
	return fmt.Sprintf("addr=%s reason=%s after-started=%t",
		observed.stopping.Addr, observed.stopping.Reason, after), nil
}

// serverStoppedEventScenario verifies the server.stopped payload and timing.
func serverStoppedEventScenario(base string) (string, error) {
	observed, err := observeServerEvents(base)
	if err != nil {
		return "", err
	}
	after := indexOf(observed.sequence, "stopped") > indexOf(observed.sequence, "stopping")
	return fmt.Sprintf("addr=%s error-empty=%t after-stopping=%t",
		observed.stopped.Addr, observed.stopped.Error == "", after), nil
}

// requestReceivedEventScenario verifies the request.received payload fields.
func requestReceivedEventScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/received", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		return nil
	}
	middleware := func(m *foundation.Middleware) {
		m.Prepend(func(engine *gin.Engine) { engine.Use(httpmiddleware.RequestID()) })
	}
	app, server, err := newHTTPServer(base, routes, middleware)
	if err != nil {
		return "", err
	}
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	var received event.RequestReceived
	bus.Listen(event.EventRequestReceived, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(event.RequestReceived); ok {
			received = value
		}
		return nil
	}))
	serveRequest(server, http.MethodGet, "/received", nil)
	return fmt.Sprintf("method=%s path=%s client-ip=%t request-id=%t received-at=%t",
		received.Method, received.Path, received.ClientIP != "", received.RequestID != "", !received.ReceivedAt.IsZero()), nil
}

// requestHandledEventScenario verifies the request.handled payload for a successful request.
func requestHandledEventScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/handled", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		return nil
	}
	app, server, err := newHTTPServer(base, routes, nil)
	if err != nil {
		return "", err
	}
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	var handled event.RequestHandled
	bus.Listen(event.EventRequestHandled, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(event.RequestHandled); ok {
			handled = value
		}
		return nil
	}))
	response := serveRequest(server, http.MethodGet, "/handled", nil)
	return fmt.Sprintf("status=%d duration-nonnegative=%t handled=%t",
		handled.Status, handled.Duration >= 0, response.Code == http.StatusOK), nil
}

// requestFailedEventScenario verifies the request.failed payload for a 5xx request.
func requestFailedEventScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/failed", func(c *gin.Context) {
			_ = c.Error(errors.New("demo request failure"))
		})
		return nil
	}
	app, server, err := newHTTPServer(base, routes, nil)
	if err != nil {
		return "", err
	}
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	var failed event.RequestFailed
	bus.Listen(event.EventRequestFailed, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		if value, ok := ev.(event.RequestFailed); ok {
			failed = value
		}
		return nil
	}))
	serveRequest(server, http.MethodGet, "/failed", nil)
	return fmt.Sprintf("status=%d method=%s path=%s error-recorded=%t",
		failed.Status, failed.Method, failed.Path, failed.Error != ""), nil
}

// requestFinishedEventScenario verifies request.finished dispatches after handled.
func requestFinishedEventScenario(base string) (string, error) {
	routes := func(_ *foundation.Application, engine *gin.Engine) error {
		engine.GET("/finished", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		return nil
	}
	app, server, err := newHTTPServer(base, routes, nil)
	if err != nil {
		return "", err
	}
	defer app.Close()

	bus, err := eventDispatcher(app)
	if err != nil {
		return "", err
	}
	log := &eventLog{}
	bus.Listen(event.EventRequestHandled, event.ListenerFunc(func(context.Context, event.Event) error {
		log.add("handled")
		return nil
	}))
	var finished event.RequestFinished
	bus.Listen(event.EventRequestFinished, event.ListenerFunc(func(_ context.Context, ev event.Event) error {
		log.add("finished")
		if value, ok := ev.(event.RequestFinished); ok {
			finished = value
		}
		return nil
	}))
	serveRequest(server, http.MethodGet, "/finished", nil)
	order := strings.Join(log.values(), ",")
	return fmt.Sprintf("status=%d after-handled=%t error-empty=%t",
		finished.Status, order == "handled,finished", finished.Error == ""), nil
}
