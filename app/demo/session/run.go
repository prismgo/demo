// Package sessiondemo contains executable examples of server-side session storage.
package sessiondemo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	configpkg "github.com/prismgo/framework/config"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/session"
)

// Result records one observable session scenario.
type Result struct {
	Case  string `json:"case"`
	Value string `json:"value"`
}

// Run executes a session catalog scenario that needs no external services.
func Run(name string) (Result, error) {
	value, err := run(name)
	if err != nil {
		return Result{}, fmt.Errorf("session demo %s: %w", name, err)
	}
	return Result{Case: name, Value: value}, nil
}

func run(name string) (string, error) {
	switch name {
	case "architecture":
		return architecture()
	case "config":
		return configFileScenario()
	case "file-driver":
		return fileDriverScenario()
	case "top-level-config", "cookie-config", "file-config", "redis-config", "lock-config":
		return configSectionScenario(name)
	case "manager-config":
		return managerConfigScenario()
	case "middleware", "recovery", "response-buffering", "store-from", "custom-manager", "flash", "now", "reflash", "keep":
		return requestScenario(name)
	case "get", "all", "subsets", "has", "exists", "missing", "put", "counters", "forget", "flush", "pull":
		return storeScenario(name)
	default:
		return runBatch(name)
	}
}

// architecture exercises every compile-facing contract named in the guide.
func architecture() (string, error) {
	var (
		_ session.Driver    = (*session.FileDriver)(nil)
		_ session.Locker    = (*session.FileDriver)(nil)
		_ session.Lock      = demoLock{}
		_ session.Encryptor = session.NopEncryptor{}
	)
	return "contracts=driver,locker,lock,encryptor", nil
}

// demoLock is a compile-facing witness for the Lock contract used by custom drivers.
type demoLock struct{}

// Release implements session.Lock.
func (demoLock) Release(context.Context) error { return nil }

// configFileScenario overrides every documented SESSION_* environment variable.
func configFileScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-config-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	lines := []string{
		"SESSION_DRIVER=redis",
		"SESSION_LIFETIME=45",
		"SESSION_EXPIRE_ON_CLOSE=true",
		"SESSION_ENCRYPT=true",
		"SESSION_ENCODING=json",
		"SESSION_CONNECTION=demo",
		"SESSION_PREFIX=demo_session",
		"SESSION_COOKIE=demo_session_cookie",
		"SESSION_PATH=/demo",
		"SESSION_DOMAIN=demo.test",
		"SESSION_SECURE_COOKIE=true",
		"SESSION_HTTP_ONLY=false",
		"SESSION_SAME_SITE=strict",
		"SESSION_FILES=" + filepath.Join(root, "sessions"),
		"SESSION_LOCK_SECONDS=7",
		"SESSION_LOCK_WAIT_SECONDS=3",
	}
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		return "", err
	}
	repo, err := configpkg.NewFromFile(envPath)
	if err != nil {
		return "", err
	}
	cfg := session.ConfigFromRepository(repo)
	parts := []string{
		fmt.Sprintf("driver=%s", cfg.Driver),
		fmt.Sprintf("lifetime=%s", cfg.Lifetime),
		fmt.Sprintf("expireOnClose=%t", cfg.ExpireOnClose),
		fmt.Sprintf("encrypt=%t", cfg.Encrypt),
		fmt.Sprintf("encoding=%s", cfg.Encoding),
		fmt.Sprintf("connection=%s", cfg.Redis.Connection),
		fmt.Sprintf("prefix=%s", cfg.Redis.Prefix),
		fmt.Sprintf("cookie=%s", cfg.Cookie.Name),
		fmt.Sprintf("path=%s", cfg.Cookie.Path),
		fmt.Sprintf("domain=%s", cfg.Cookie.Domain),
		fmt.Sprintf("secure=%t", cfg.Cookie.Secure),
		fmt.Sprintf("httpOnly=%t", cfg.Cookie.HTTPOnly),
		fmt.Sprintf("sameSite=%s", cfg.Cookie.SameSite),
		fmt.Sprintf("lockTTL=%s", cfg.Lock.TTL),
		fmt.Sprintf("lockWait=%s", cfg.Lock.Wait),
	}
	return strings.Join(parts, " "), nil
}

// fileDriverScenario persists a session through the file driver and reloads it.
func fileDriverScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-file-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Files = root
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	store := freshStore(manager)
	store.Put("user_id", int64(1001))
	if err := store.Save(context.Background()); err != nil {
		return "", err
	}
	info, err := os.Stat(filepath.Join(root, store.ID()))
	if err != nil {
		return "", err
	}
	restored, err := restoreSession(manager, store.ID())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file=true mode=%v restored=%v", info.Mode().Perm(), restored.Get("user_id")), nil
}

// configSectionScenario verifies the documented defaults for one configuration table.
func configSectionScenario(name string) (string, error) {
	cfg := session.ConfigFromRepository(nil)
	switch name {
	case "top-level-config":
		return fmt.Sprintf("driver=%s lifetime=%s expireOnClose=%t encrypt=%t encoding=%s", cfg.Driver, cfg.Lifetime, cfg.ExpireOnClose, cfg.Encrypt, encodingOf(cfg)), nil
	case "cookie-config":
		return fmt.Sprintf("cookie=%s path=%s domain=%q secure=%t httpOnly=%t sameSite=%s", cfg.Cookie.Name, cfg.Cookie.Path, cfg.Cookie.Domain, cfg.Cookie.Secure, cfg.Cookie.HTTPOnly, cfg.Cookie.SameSite), nil
	case "file-config":
		return fmt.Sprintf("files=%s", cfg.Files), nil
	case "redis-config":
		return fmt.Sprintf("connection=%s prefix=%s", cfg.Redis.Connection, cfg.Redis.Prefix), nil
	case "lock-config":
		return fmt.Sprintf("lockSeconds=%s lockWait=%s", cfg.Lock.TTL, cfg.Lock.Wait), nil
	}
	return "", fmt.Errorf("unknown config section %q", name)
}

// encodingOf reports the codec the manager will resolve for the given config.
func encodingOf(cfg session.Config) string {
	if cfg.Encoding == "" {
		return "msgpack"
	}
	return cfg.Encoding
}

// managerConfigScenario constructs a manager from explicit config.
func managerConfigScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-manager-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Lifetime = 30 * time.Minute
	cfg.Cookie.Name = "admin_session"
	cfg.Cookie.Secure = true
	cfg.Cookie.SameSite = "lax"
	cfg.Files = root
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	got := manager.Config()
	_, invalidErr := session.NewManager(session.Config{Encoding: "yaml"}, nil)
	return fmt.Sprintf("lifetime=%s cookie=%s secure=%t files=true invalid-encoding=%t", got.Lifetime, got.Cookie.Name, got.Cookie.Secure, invalidErr != nil), nil
}

// storeScenario exercises request-scoped Store read/write APIs through a real manager.
func storeScenario(name string) (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	store := freshStore(manager)
	switch name {
	case "get":
		store.Put("name", "alice")
		return fmt.Sprintf("value=%v default=%v absent=%v", store.Get("name", "guest"), store.Get("missing", "guest"), store.Get("missing")), nil
	case "all":
		store.Put("a", "1")
		store.Put("b", "2")
		values := store.All()
		delete(values, "a")
		return fmt.Sprintf("count=%d isolated=%t", len(store.All()), len(values) == 1), nil
	case "subsets":
		store.Put("a", "1")
		store.Put("b", "2")
		store.Put("c", "3")
		return fmt.Sprintf("only=%d except=%d", len(store.Only("a", "b", "missing")), len(store.Except("c"))), nil
	case "has":
		store.Put("present", "x")
		store.Put("empty", nil)
		return fmt.Sprintf("has=%t nil=%t", store.Has("present"), store.Has("empty")), nil
	case "exists":
		store.Put("empty", nil)
		return fmt.Sprintf("exists=%t missing=%t", store.Exists("empty"), store.Exists("nope")), nil
	case "missing":
		store.Put("present", "x")
		return fmt.Sprintf("missing=%t present=%t", store.Missing("nope"), store.Missing("present")), nil
	case "put":
		store.Put("user_id", int64(1001))
		store.Put("filters", map[string]any{"status": "open", "page": 1})
		store.Put("tags", []string{"a", "b"})
		if err := store.Save(context.Background()); err != nil {
			return "", err
		}
		restored, err := restoreSession(manager, store.ID())
		if err != nil {
			return "", err
		}
		filters, mapOK := restored.Get("filters").(map[string]any)
		tags, sliceOK := restored.Get("tags").([]any)
		return fmt.Sprintf("scalar=%v map=%t slice=%t", restored.Get("user_id"), mapOK && filters["status"] == "open", sliceOK && len(tags) == 2), nil
	case "counters":
		store.Put("retry_count", int64(0))
		first, err := store.Increment("retry_count")
		if err != nil {
			return "", err
		}
		second, err := store.Increment("retry_count", 2)
		if err != nil {
			return "", err
		}
		remaining, err := store.Decrement("quota", 2)
		if err != nil {
			return "", err
		}
		store.Put("bad", "text")
		_, invalidErr := store.Increment("bad")
		return fmt.Sprintf("first=%d second=%d remaining=%d invalid=%t", first, second, remaining, errors.Is(invalidErr, session.ErrInvalidValueType)), nil
	case "forget":
		store.Put("draft", "d")
		store.Put("notice", "n")
		store.Forget("draft", "notice")
		return fmt.Sprintf("gone=%t gone=%t", store.Missing("draft"), store.Missing("notice")), nil
	case "flush":
		store.Put("user_id", int64(7))
		store.Flash("status", "ok")
		beforeID := store.ID()
		store.Flush()
		return fmt.Sprintf("data=%d flash=%v id-same=%t", len(store.All()), store.Get("status"), store.ID() == beforeID), nil
	case "pull":
		store.Flash("notice", "hello")
		pulled := store.Pull("notice", "")
		fallback := store.Pull("missing", "fallback")
		if err := store.Save(context.Background()); err != nil {
			return "", err
		}
		restored, err := restoreSession(manager, store.ID())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("pulled=%v exists=%t fallback=%v", pulled, restored.Exists("notice"), fallback), nil
	}
	return "", fmt.Errorf("unknown store scenario %q", name)
}

// tempManager builds a file-backed manager rooted in a throwaway directory.
func tempManager() (*session.Manager, func(), error) {
	root, err := os.MkdirTemp("", "prismgo-session-demo-")
	if err != nil {
		return nil, nil, err
	}
	cfg := session.DefaultConfig()
	cfg.Files = root
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		os.RemoveAll(root)
		return nil, nil, err
	}
	return manager, func() { os.RemoveAll(root) }, nil
}

// freshStore starts a request without any session cookie.
func freshStore(manager *session.Manager) *session.Store {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	store, err := manager.Start(context.Background(), r, httptest.NewRecorder())
	if err != nil {
		panic(err)
	}
	return store
}

// restoreSession starts a request that carries the given session ID cookie.
func restoreSession(manager *session.Manager, id string) (*session.Store, error) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: session.DefaultCookieName, Value: id})
	return manager.Start(context.Background(), r, httptest.NewRecorder())
}

// firstCookie locates the named cookie in a response.
func firstCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// bodyText drains a response body for scenario assertions.
func bodyText(response *http.Response) string {
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

// tempManagerWithRoot builds a file manager and also returns the backing directory.
func tempManagerWithRoot() (*session.Manager, string, func(), error) {
	root, err := os.MkdirTemp("", "prismgo-session-recovery-")
	if err != nil {
		return nil, "", nil, err
	}
	cfg := session.DefaultConfig()
	cfg.Files = root
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		os.RemoveAll(root)
		return nil, "", nil, err
	}
	return manager, root, func() { os.RemoveAll(root) }, nil
}

// newDemoSessionID mints a syntactically valid opaque session ID for recovery probes.
func newDemoSessionID() string {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// withGinTestMode runs fn with Gin's test mode forced on.
func withGinTestMode(fn func() (string, error)) (string, error) {
	previous := gin.Mode()
	gin.SetMode(gin.TestMode)
	defer gin.SetMode(previous)
	return fn()
}

// sessionEngine mounts a StartSession route backed by the given manager.
func sessionEngine(manager *session.Manager, routes func(*gin.RouterGroup)) *gin.Engine {
	engine := gin.New()
	group := engine.Group("/")
	group.Use(httpmiddleware.StartSession(session.WithManager(manager)))
	routes(group)
	return engine
}

// roundTrip performs one request through the engine, carrying session cookies.
func roundTrip(engine *gin.Engine, jar []*http.Cookie, method string, path string) (*http.Response, []*http.Cookie, error) {
	r := httptest.NewRequest(method, path, nil)
	for _, c := range jar {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, r)
	response := w.Result()
	if response.StatusCode >= 500 {
		return response, response.Cookies(), fmt.Errorf("session request %s %s failed with status %d", method, path, response.StatusCode)
	}
	return response, response.Cookies(), nil
}

// carryCookies merges a fresh Set-Cookie list into the jar, replacing by name+scope.
func carryCookies(jar []*http.Cookie, fresh []*http.Cookie) []*http.Cookie {
	merged := append([]*http.Cookie{}, jar...)
	for _, c := range fresh {
		replaced := false
		for i, old := range merged {
			if old.Name == c.Name && old.Path == c.Path && old.Domain == c.Domain {
				merged[i] = c
				replaced = true
				break
			}
		}
		if !replaced {
			merged = append(merged, c)
		}
	}
	return merged
}

// setScenarioEnv applies environment variables for one scenario and returns a restore closure.
func setScenarioEnv(values map[string]string) func() {
	previous := make(map[string]*string, len(values))
	for key, value := range values {
		if old, ok := os.LookupEnv(key); ok {
			copied := old
			previous[key] = &copied
		} else {
			previous[key] = nil
		}
		_ = os.Setenv(key, value)
	}
	return func() {
		for key, old := range previous {
			if old == nil {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, *old)
		}
	}
}
