// Package cookiedemo contains executable examples of Cookie values and request-scoped writes.
package cookiedemo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/container"
	containercontract "github.com/prismgo/framework/contracts/container"
	"github.com/prismgo/framework/cookie"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/session"
)

// Result records one observable Cookie scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a Cookie catalog scenario.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("cookie demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(name string) (string, error) {
	switch name {
	case "value-object":
		c := cookie.New("theme", "dark", 10)
		return fmt.Sprintf("name=%s value=%s minutes=%d", c.Name, c.Value, c.Minutes), nil
	case "scope-deduplication":
		q := cookie.NewQueue()
		q.Make("theme", "first", 5)
		q.Make("theme", "last", 5)
		q.Make("theme", "admin", 5, cookie.Path("/admin"))
		w := httptest.NewRecorder()
		if err := q.Flush(w); err != nil {
			return "", err
		}
		return fmt.Sprintf("headers=%d values=%s", len(w.Result().Cookies()), cookieValues(w.Result().Cookies())), nil
	case "defaults":
		c := cookie.New("theme", "dark", cookie.ForeverMinutes)
		return fmt.Sprintf("path=%s httpOnly=%t secure=%t sameSite=%q forever=%d", c.Path, c.HTTPOnly, c.Secure, c.SameSite, c.Minutes), nil
	case "provider":
		return providerScenario()
	case "middleware", "flush-failure", "queue-make", "queue-forever", "queue-cookie", "queue-from":
		return requestScenario(name)
	case "session-queue":
		return sessionQueue()
	case "make":
		a := cookie.New("theme", "dark", 5)
		b := cookie.Make("theme", "dark", 5)
		return fmt.Sprintf("equal=%t name=%s value=%s", a == b, a.Name, a.Value), nil
	case "forever":
		c := cookie.Forever("theme", "dark")
		return fmt.Sprintf("minutes=%d value=%s", c.Minutes, c.Value), nil
	case "scope-options":
		c := cookie.New("theme", "dark", 5, cookie.Path("/admin"), cookie.Domain("example.test"))
		return fmt.Sprintf("path=%s domain=%s", c.Path, c.Domain), nil
	case "security-flags":
		c := cookie.New("theme", "dark", 5, cookie.HTTPOnly(false), cookie.Secure(true))
		return fmt.Sprintf("httpOnly=%t secure=%t", c.HTTPOnly, c.Secure), nil
	case "raw":
		c := cookie.New("theme", "dark", 5, cookie.Raw(true))
		return fmt.Sprintf("raw=%t value=%s", c.Raw, c.Value), nil
	case "scope-option":
		scope := cookie.Scope{Name: "ignored", Path: "/admin", Domain: "example.test"}
		a := cookie.Make("one", "a", 5, cookie.ScopeOption(scope))
		b := cookie.Make("two", "b", 5, cookie.ScopeOption(scope))
		return fmt.Sprintf("names=%s,%s path=%s domain=%s", a.Name, b.Name, a.Path, a.Domain), nil
	case "minutes":
		a := cookie.Make("short", "a", 5)
		b := cookie.Make("session", "b", 0)
		short, err := a.ToHTTP()
		if err != nil {
			return "", err
		}
		session, err := b.ToHTTP()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("minutes=%d,%d maxAge=%d,%d persistent=%t,%t", a.Minutes, b.Minutes, short.MaxAge, session.MaxAge, !short.Expires.IsZero(), !session.Expires.IsZero()), nil
	case "expires-at":
		now := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
		c := cookie.Make("theme", "dark", 5, cookie.ExpiresAt(now.Add(time.Hour)))
		h, err := c.ToHTTP()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("expires=%s maxAge=%d", h.Expires.UTC().Format(time.RFC3339), h.MaxAge), nil
	case "max-age":
		now := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
		c := cookie.Make("theme", "dark", 0, cookie.ExpiresAt(now), cookie.MaxAge(42))
		h, err := c.ToHTTP()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("expires=%s maxAge=%d", h.Expires.UTC().Format(time.RFC3339), h.MaxAge), nil
	case "attach":
		w := httptest.NewRecorder()
		if err := cookie.New("first", "one", 0).Attach(w); err != nil {
			return "", err
		}
		if err := cookie.Attach(w, cookie.New("second", "two", 0)); err != nil {
			return "", err
		}
		return cookieValues(w.Result().Cookies()), nil
	case "invalid-name":
		w := httptest.NewRecorder()
		err := cookie.New("bad name", "secret", 0).Attach(w)
		return fmt.Sprintf("invalid=%t headers=%d", errors.Is(err, cookie.ErrInvalidCookieName), len(w.Result().Cookies())), nil
	case "attach-context":
		ctx := context.WithValue(context.Background(), contextKey{}, "request-42")
		w := httptest.NewRecorder()
		err := cookie.Make("trace", "value", 0).Attach(w, cookie.WithContext(ctx), cookie.WithSigner(contextSigner{}))
		if err != nil {
			return "", err
		}
		return cookieValues(w.Result().Cookies()), nil
	case "attach-now":
		now := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
		w := httptest.NewRecorder()
		if err := cookie.Make("theme", "dark", 5).Attach(w, cookie.WithNow(now)); err != nil {
			return "", err
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			return "", fmt.Errorf("attached cookies = %d, want 1", len(cookies))
		}
		h := cookies[0]
		return fmt.Sprintf("expires=%s maxAge=%d", h.Expires.UTC().Format(time.RFC3339), h.MaxAge), nil
	case "queued", "has-queued", "scoped-queued", "unqueue":
		return queuedScenario(name)
	case "queue-not-found":
		_, err := cookie.QueueMakeFrom(nil, "theme", "dark", 5)
		return fmt.Sprintf("missing=%t", errors.Is(err, cookie.ErrQueueNotFound)), nil
	case "process-queue":
		cookie.QueueMake("theme", "dark", 5)
		c, ok := cookie.Queued("theme")
		cookie.Unqueue("theme")
		return fmt.Sprintf("queued=%t value=%s removed=%t", ok, c.Value, !cookie.HasQueued("theme")), nil
	default:
		return runRemaining(name)
	}
}

func providerScenario() (string, error) {
	registry := container.NewContainer()
	app := providerApplication{registry: registry}
	if err := (cookie.ServiceProvider{}).Register(app); err != nil {
		return "", fmt.Errorf("register cookie provider: %w", err)
	}
	before := registry.Bound("cookie.queue") && !registry.Resolved("cookie.queue")
	raw, err := registry.Make("cookie.queue")
	if err != nil {
		return "", fmt.Errorf("resolve cookie queue: %w", err)
	}
	_, ok := raw.(*cookie.Queue)
	return fmt.Sprintf("lazy=%t resolved=%t queue=%t", before, registry.Resolved("cookie.queue"), ok), nil
}

type providerApplication struct{ registry containercontract.Container }

func (a providerApplication) Container() containercontract.Container { return a.registry }

func cookieValues(cookies []*http.Cookie) string {
	values := make([]string, 0, len(cookies))
	for _, c := range cookies {
		values = append(values, c.Name+"="+c.Value+"@"+c.Path)
	}
	return strings.Join(values, ",")
}

type contextKey struct{}
type contextSigner struct{}

func (contextSigner) Sign(ctx context.Context, _ string, value string) (string, error) {
	trace, _ := ctx.Value(contextKey{}).(string)
	return value + ":" + trace, nil
}
func (contextSigner) Unsign(_ context.Context, _ string, value string) (string, error) {
	return value, nil
}

func queuedScenario(name string) (string, error) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(previousMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(cookie.QueueKey, cookie.NewQueue())
	q, _ := cookie.QueueFrom(c)
	q.Make("theme", "dark", 5)
	switch name {
	case "queued":
		got, ok, err := cookie.QueuedFrom(c, "theme")
		return fmt.Sprintf("found=%t value=%s", ok, got.Value), err
	case "has-queued":
		found, err := cookie.HasQueuedFrom(c, "theme")
		return fmt.Sprintf("found=%t", found), err
	case "scoped-queued":
		q.Make("theme", "admin", 5, cookie.Path("/admin"), cookie.Domain("example.test"))
		got, ok, err := cookie.QueuedFrom(c, "theme", cookie.Scope{Path: "/admin", Domain: "example.test"})
		if err != nil {
			return "", err
		}
		defaultCookie, defaultFound, err := cookie.QueuedFrom(c, "theme")
		if err != nil {
			return "", err
		}
		_, otherDomain, err := cookie.QueuedFrom(c, "theme", cookie.Scope{Path: "/admin", Domain: "other.test"})
		return fmt.Sprintf("found=%t value=%s default=%t:%s other-domain=%t", ok, got.Value, defaultFound, defaultCookie.Value, otherDomain), err
	case "unqueue":
		if err := cookie.UnqueueFrom(c, "theme"); err != nil {
			return "", err
		}
		return fmt.Sprintf("remaining=%t", q.HasQueued("theme")), nil
	default:
		return "", fmt.Errorf("unknown queued scenario %q", name)
	}
}

func requestScenario(name string) (string, error) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(previousMode)
	engine := gin.New()
	engine.Use(httpmiddleware.QueuedCookies())
	var scenarioErr error
	engine.GET("/", func(c *gin.Context) {
		switch name {
		case "middleware":
			_, scenarioErr = cookie.QueueMakeFrom(c, "theme", "dark", 5)
		case "flush-failure":
			_, scenarioErr = cookie.QueueMakeFrom(c, "bad name", "secret", 5)
		case "queue-make":
			_, scenarioErr = cookie.QueueMakeFrom(c, "theme", "dark", 5)
		case "queue-forever":
			_, scenarioErr = cookie.QueueForeverFrom(c, "theme", "dark")
		case "queue-cookie":
			scenarioErr = cookie.QueueCookieFrom(c, cookie.Make("theme", "dark", 5))
		case "queue-from":
			q, ok := cookie.QueueFrom(c)
			if !ok {
				scenarioErr = cookie.ErrQueueNotFound
				break
			}
			q.Make("theme", "dark", 5)
		}
		c.String(http.StatusCreated, "queued")
	})
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if scenarioErr != nil {
		return "", scenarioErr
	}
	if name == "queue-forever" {
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			return "", fmt.Errorf("queued forever cookies = %d, want 1", len(cookies))
		}
		return fmt.Sprintf("status=%d cookies=%s maxAge=%d", w.Code, cookieValues(cookies), cookies[0].MaxAge), nil
	}
	return fmt.Sprintf("status=%d cookies=%s", w.Code, cookieValues(w.Result().Cookies())), nil
}

func sessionQueue() (value string, err error) {
	root, err := os.MkdirTemp("", "prismgo-cookie-session-")
	if err != nil {
		return "", fmt.Errorf("create session directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(root); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("remove session directory: %w", removeErr))
		}
	}()
	cfg := session.DefaultConfig()
	cfg.Files = root
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", fmt.Errorf("create session manager: %w", err)
	}
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(previousMode)
	engine := gin.New()
	engine.Use(httpmiddleware.StartSession(session.WithManager(manager)))
	var scenarioErr error
	engine.GET("/", func(c *gin.Context) {
		_, scenarioErr = cookie.QueueMakeFrom(c, "theme", "dark", 5)
		c.String(http.StatusOK, "session")
	})
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if scenarioErr != nil {
		return "", scenarioErr
	}
	return fmt.Sprintf("status=%d cookies=%s", w.Code, cookieValues(w.Result().Cookies())), nil
}
