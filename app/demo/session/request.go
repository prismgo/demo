package sessiondemo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/foundation"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/session"

	// Load the demo's configuration defaults before booting the default manager.
	_ "prismgo-demo/config"
)

func requestScenario(name string) (string, error) {
	return withGinTestMode(func() (string, error) {
		switch name {
		case "middleware":
			return middlewareScenario()
		case "recovery":
			return recoveryScenario()
		case "response-buffering":
			return responseBufferingScenario()
		case "store-from":
			return storeFromScenario()
		case "custom-manager":
			return customManagerScenario()
		case "flash", "now", "reflash", "keep":
			return flashScenario(name)
		}
		return "", fmt.Errorf("unknown request scenario %q", name)
	})
}

// middlewareScenario walks the documented StartSession request flow end to end.
func middlewareScenario() (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	engine := sessionEngine(manager, func(group *gin.RouterGroup) {
		group.GET("/write", func(c *gin.Context) {
			_ = session.Put(c, "user_id", int64(1001))
			_ = session.Flash(c, "notice", "saved")
			c.String(http.StatusOK, "written")
		})
		group.GET("/read", func(c *gin.Context) {
			userID := session.Get(c, "user_id", int64(0))
			notice := session.Pull(c, "notice", "")
			c.String(http.StatusOK, "user=%v notice=%v", userID, notice)
		})
	})
	first, cookies, err := roundTrip(engine, nil, http.MethodGet, "/write")
	if err != nil {
		return "", err
	}
	if first.StatusCode != http.StatusOK {
		return "", fmt.Errorf("middleware scenario write status = %d, want 200", first.StatusCode)
	}
	idCookie := firstCookie(cookies, session.DefaultCookieName)
	if idCookie == nil {
		return "", errors.New("middleware scenario: missing session ID cookie")
	}
	second, _, err := roundTrip(engine, cookies, http.MethodGet, "/read")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("cookie=%s httpOnly=%t sameSiteLax=%t persistent=%t read=%q",
		idCookie.Name, idCookie.HttpOnly, idCookie.SameSite == http.SameSiteLaxMode, idCookie.MaxAge > 0, bodyText(second)), nil
}

// recoveryScenario verifies missing, expired, corrupted, and invalid IDs start fresh sessions.
func recoveryScenario() (string, error) {
	manager, root, cleanup, err := tempManagerWithRoot()
	if err != nil {
		return "", err
	}
	defer cleanup()
	fresh := func(id string) (bool, error) {
		store, err := restoreSession(manager, id)
		if err != nil {
			return false, err
		}
		return store.ID() != id, nil
	}
	missing, err := fresh(newDemoSessionID())
	if err != nil {
		return "", err
	}
	expiredID := newDemoSessionID()
	driver, err := session.ResolveDriver("file", manager.Config())
	if err != nil {
		return "", err
	}
	now := time.Now()
	until := now.Add(time.Second)
	if err := driver.Write(context.Background(), expiredID, session.Payload{
		ID: expiredID, Values: map[string]any{"user_id": int64(1)},
		CreatedAt: now, LastActivity: now,
	}, &until); err != nil {
		return "", err
	}
	time.Sleep(1100 * time.Millisecond)
	expired, err := fresh(expiredID)
	if err != nil {
		return "", err
	}
	corruptedID := newDemoSessionID()
	if err := os.WriteFile(filepath.Join(root, corruptedID), []byte("{{{not-a-session-payload"), 0o600); err != nil {
		return "", err
	}
	corrupted, err := fresh(corruptedID)
	if err != nil {
		return "", err
	}
	invalid, err := fresh("short-id")
	if err != nil {
		return "", err
	}
	if _, statErr := os.Stat(filepath.Join(root, expiredID)); !errors.Is(statErr, os.ErrNotExist) {
		return "", fmt.Errorf("expired session file = stat error %v, want removed", statErr)
	}
	return fmt.Sprintf("missing=%t expired=%t corrupted=%t invalid=%t", missing, expired, corrupted, invalid), nil
}

// responseBufferingScenario shows the buffered commit and its streaming-route limitation.
func responseBufferingScenario() (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	blockedDir, err := os.MkdirTemp("", "prismgo-session-blocked-")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = os.Chmod(blockedDir, 0o700)
		_ = os.RemoveAll(blockedDir)
	}()
	blockedCfg := session.DefaultConfig()
	blockedCfg.Files = blockedDir
	blockedManager, err := session.NewManager(blockedCfg, nil)
	if err != nil {
		return "", err
	}
	// 目录改为只读后，Save 的原子写入失败，验证 commit 钩子出错时缓冲响应被整体丢弃。
	if err := os.Chmod(blockedDir, 0o500); err != nil {
		return "", err
	}
	engine := sessionEngine(manager, func(group *gin.RouterGroup) {
		group.GET("/ok", func(c *gin.Context) {
			_, _ = c.Writer.Write([]byte("chunk-1"))
			_ = session.Put(c, "user_id", int64(1))
			c.String(http.StatusOK, "|chunk-2")
		})
	})
	blockedEngine := sessionEngine(blockedManager, func(group *gin.RouterGroup) {
		group.GET("/save-fails", func(c *gin.Context) {
			_ = session.Put(c, "user_id", int64(1))
			c.String(http.StatusOK, "secret-body")
		})
	})
	okResponse, okCookies, err := roundTrip(engine, nil, http.MethodGet, "/ok")
	if err != nil {
		return "", err
	}
	badRequest := httptest.NewRequest(http.MethodGet, "/save-fails", nil)
	badRecorder := httptest.NewRecorder()
	blockedEngine.ServeHTTP(badRecorder, badRequest)
	badResponse := badRecorder.Result()
	suppressed := bodyText(badResponse) == ""
	if len(okCookies) != 1 || firstCookie(okCookies, session.DefaultCookieName) == nil {
		return "", fmt.Errorf("buffering scenario cookies = %d, want 1 session cookie", len(okCookies))
	}
	if badResponse.StatusCode != http.StatusInternalServerError {
		return "", fmt.Errorf("save-failure status = %d, want 500", badResponse.StatusCode)
	}
	if !suppressed {
		return "", errors.New("save-failure scenario: buffered body leaked despite failed commit")
	}
	return fmt.Sprintf("delivered=%q save-failed-status=%d body-suppressed=%t",
		bodyText(okResponse), badResponse.StatusCode, suppressed), nil
}

// storeFromScenario exercises StoreFrom and the package facade shortcuts.
func storeFromScenario() (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	engine := gin.New()
	bare := engine.Group("/bare")
	bare.GET("/read", func(c *gin.Context) {
		value := session.Get(c, "user_id", "guest")
		_ = session.Put(c, "user_id", int64(1))
		_, found := session.StoreFrom(c)
		c.String(http.StatusOK, "default=%v found=%t put=%t", value, found, errors.Is(session.Put(c, "x", 1), session.ErrInvalidConfig))
	})
	guarded := engine.Group("/", httpmiddleware.StartSession(session.WithManager(manager)))
	guarded.GET("/read", func(c *gin.Context) {
		store, ok := session.StoreFrom(c)
		if !ok {
			c.String(http.StatusInternalServerError, "no store")
			return
		}
		_ = session.Put(c, "user_id", int64(1001))
		c.String(http.StatusOK, "id-len=%d facade-get=%v", len(store.ID()), session.Get(c, "user_id"))
	})
	storeResponse, _, err := roundTrip(engine, nil, http.MethodGet, "/read")
	if err != nil {
		return "", err
	}
	bareResponse, _, err := roundTrip(engine, nil, http.MethodGet, "/bare/read")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("store=[%s] bare=[%s]", bodyText(storeResponse), bodyText(bareResponse)), nil
}

// customManagerScenario injects a dedicated manager and checks default resolution.
func customManagerScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-session-custom-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(root)
	cfg := session.DefaultConfig()
	cfg.Cookie.Name = "admin_session"
	cfg.Files = root
	manager, err := session.NewManager(cfg, nil)
	if err != nil {
		return "", err
	}
	engine := gin.New()
	admin := engine.Group("/", httpmiddleware.StartSession(session.WithManager(manager)))
	admin.GET("/visit", func(c *gin.Context) {
		if err := session.Put(c, "scope", "admin"); err != nil {
			c.String(http.StatusInternalServerError, "%v", err)
			return
		}
		c.String(http.StatusOK, "scope=%v", session.Get(c, "scope"))
	})
	first, cookies, err := roundTrip(engine, nil, http.MethodGet, "/visit")
	if err != nil {
		return "", err
	}
	if firstCookie(cookies, "admin_session") == nil {
		return "", errors.New("custom manager scenario: admin_session cookie missing")
	}
	second, _, err := roundTrip(engine, cookies, http.MethodGet, "/visit")
	if err != nil {
		return "", err
	}
	defaultResult, err := defaultManagerScenario()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("cookie=admin_session r1=%q r2=%q default=%s",
		bodyText(first), bodyText(second), defaultResult), nil
}

// defaultManagerScenario boots the real application so StartSession can resolve
// session.Default() through the container facade instead of WithManager.
func defaultManagerScenario() (string, error) {
	basePath, err := os.MkdirTemp("", "prismgo-session-default-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(basePath)
	restore := setScenarioEnv(map[string]string{
		"APP_ENV": "testing", "APP_KEY": "base64:ZGVtby10ZXN0LWtleS1kbz1ub3QtdXNlLWluLXByb2R1Y3Rpb24=",
		"DB_CONNECTION": "sqlite", "DB_DATABASE": filepath.Join(basePath, "database.sqlite"),
		"CACHE_STORE": "memory", "QUEUE_CONNECTION": "sync",
		"SESSION_DRIVER": "file", "SESSION_FILES": filepath.Join(basePath, "storage", "framework", "sessions"),
	})
	defer restore()
	app := foundation.Configure(basePath).Create()
	if err := app.Boot(); err != nil {
		return "", fmt.Errorf("boot default manager application: %w", err)
	}
	defer func() {
		_ = app.CloseContext(context.Background())
	}()
	if session.Resolve() == nil {
		return "", errors.New("default session manager did not resolve")
	}
	engine := gin.New()
	group := engine.Group("/", httpmiddleware.StartSession())
	group.GET("/visit", func(c *gin.Context) {
		_ = session.Put(c, "source", "default-manager")
		c.String(http.StatusOK, "source=%v", session.Get(c, "source"))
	})
	first, cookies, err := roundTrip(engine, nil, http.MethodGet, "/visit")
	if err != nil {
		return "", err
	}
	if first.StatusCode != http.StatusOK {
		return "", fmt.Errorf("default manager scenario first status = %d, want 200", first.StatusCode)
	}
	if firstCookie(cookies, session.DefaultCookieName) == nil {
		return "", errors.New("default manager scenario: prismgo_session cookie missing")
	}
	second, _, err := roundTrip(engine, cookies, http.MethodGet, "/visit")
	if err != nil {
		return "", err
	}
	if body := bodyText(second); body != "source=default-manager" {
		return "", fmt.Errorf("default manager body = %q, want source=default-manager", body)
	}
	return "resolved=true persisted=true", nil
}

// flashScenario walks the documented flash lifecycle across requests.
//
// 每个 case 走固定的请求序列，并在每一步记录可见的 flash key 集合与已消失的 key，
// 直接对应文档「当前可读 → 下一次可读 → 再下一次清理」的生命周期表。
func flashScenario(name string) (string, error) {
	manager, cleanup, err := tempManager()
	if err != nil {
		return "", err
	}
	defer cleanup()
	key := map[string][]string{"flash": {"status"}, "now": {"preview"}, "reflash": {"a", "b"}, "keep": {"kept", "dropped"}}[name]
	engine := sessionEngine(manager, func(group *gin.RouterGroup) {
		group.GET("/walk", func(c *gin.Context) {
			switch c.Request.URL.Query().Get("step") {
			case "r1":
				switch name {
				case "flash":
					_ = session.Flash(c, "status", "ok")
				case "now":
					_ = session.Now(c, "preview", "once")
				case "reflash":
					_ = session.Flash(c, "a", "one")
					_ = session.Flash(c, "b", "two")
				case "keep":
					_ = session.Flash(c, "kept", "v1")
					_ = session.Flash(c, "dropped", "v2")
				}
			case "r2":
				switch name {
				case "reflash":
					_ = session.Reflash(c)
				case "keep":
					_ = session.Keep(c, "kept")
				}
			}
			visible := make([]string, 0, len(key))
			for _, k := range key {
				if session.Exists(c, k) {
					visible = append(visible, fmt.Sprintf("%s=%v", k, session.Get(c, k)))
				} else {
					visible = append(visible, k+"=-")
				}
			}
			c.String(http.StatusOK, "%s", strings.Join(visible, " "))
		})
	})
	var jar []*http.Cookie
	results := make([]string, 0, 4)
	for _, step := range flashPlan(name) {
		r := httptest.NewRequest(http.MethodGet, "/walk?step="+step, nil)
		for _, c := range jar {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, r)
		response := w.Result()
		if response.StatusCode != http.StatusOK {
			return "", fmt.Errorf("flash scenario %s %s status = %d, want 200", name, step, response.StatusCode)
		}
		jar = carryCookies(jar, response.Cookies())
		results = append(results, step+":"+bodyText(response))
	}
	return name + " " + strings.Join(results, " "), nil
}

// flashPlan returns the request sequence that distinguishes each flash API.
func flashPlan(name string) []string {
	switch name {
	case "now":
		return []string{"r1", "r2"}
	case "flash":
		return []string{"r1", "r2", "r3"}
	default:
		return []string{"r1", "r2", "r3", "r4"}
	}
}
