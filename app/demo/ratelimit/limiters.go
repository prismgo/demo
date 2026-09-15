package ratelimitdemo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prismgo/framework/cache"
	configpkg "github.com/prismgo/framework/config"
	httpmiddleware "github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/ratelimit"
)

// autoInitializationScenario resolves the application's pre-initialized global limiter.
func autoInitializationScenario() (string, error) {
	limiter := ratelimit.Resolve()
	if limiter == nil {
		return "", fmt.Errorf("global ratelimit limiter was not initialized")
	}
	driver := strings.TrimSpace(configpkg.GetString("cache.limiter.driver", ""))
	if driver == "" {
		driver = cache.DefaultName()
	}
	ctx := context.Background()
	if _, err := limiter.Hit(ctx, "auto-initialization", time.Minute); err != nil {
		return "", fmt.Errorf("hit global limiter: %w", err)
	}
	attempts, err := limiter.Attempts(ctx, "auto-initialization")
	if err != nil {
		return "", fmt.Errorf("read global attempts: %w", err)
	}
	return fmt.Sprintf("resolved=true store=%s attempts=%d", driver, attempts), nil
}

// quickStartScenario registers a named limiter and mounts it on a login route.
func quickStartScenario() (string, error) {
	ratelimit.For("login", func(c *gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(5).By(c.ClientIP())}
	})

	engine := gin.New()
	engine.POST("/auth/login", httpmiddleware.Throttle("login"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	allowed := 0
	blocked := 0
	for range 6 {
		if perform(engine, http.MethodPost, "/auth/login").Code == http.StatusOK {
			allowed++
			continue
		}
		blocked++
	}
	return fmt.Sprintf("allowed=%d blocked=%d", allowed, blocked), nil
}

// explicitLimiterScenario builds a limiter from an explicit cache manager without an application.
func explicitLimiterScenario() (string, error) {
	manager, err := cache.NewManager(cache.Config{
		Default: "memory",
		Prefix:  "ratelimit_demo",
		Stores:  map[string]cache.StoreConfig{"memory": {Driver: "memory"}},
	})
	if err != nil {
		return "", fmt.Errorf("create explicit cache manager: %w", err)
	}
	defer func() { _ = manager.Close() }()

	limiter := ratelimit.New(manager.Default())
	limiter.For("standalone", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(2)}
	})
	engine := gin.New()
	engine.GET("/standalone", httpmiddleware.ThrottleFor(limiter, "standalone"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	status := perform(engine, http.MethodGet, "/standalone").Code
	return fmt.Sprintf("manager=memory registered=%t status=%d", limiter.Limiter("standalone") != nil, status), nil
}

// namedRegistrationScenario registers and reads back a named limiter.
func namedRegistrationScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("login", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(5).By("ip:127.0.0.1")}
	})
	resolver := limiter.Limiter("login")
	if resolver == nil {
		return "", fmt.Errorf("named limiter login is not registered")
	}
	limits := resolver(testContext())
	if len(limits) != 1 {
		return "", fmt.Errorf("named limiter login limits = %d, want 1", len(limits))
	}
	limit := limits[0]
	return fmt.Sprintf("registered=true max=%d decay=%s key=%s", limit.MaxAttempts, limit.Decay, limit.Key), nil
}

// namedLookupScenario distinguishes registered, missing, and blank limiter names.
func namedLookupScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("api", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(60)}
	})
	return fmt.Sprintf("found=%t missing=%t blank=%t",
		limiter.Limiter("api") != nil,
		limiter.Limiter("missing") == nil,
		limiter.Limiter("  ") == nil), nil
}

// unregisteredPassThroughScenario lets requests through when the named limiter is absent.
func unregisteredPassThroughScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	engine := gin.New()
	engine.GET("/open", httpmiddleware.ThrottleFor(limiter, "ghost"), func(c *gin.Context) {
		c.String(http.StatusOK, "open")
	})
	recorder := perform(engine, http.MethodGet, "/open")
	return fmt.Sprintf("resolver-nil=%t status=%d body=%s",
		limiter.Limiter("ghost") == nil, recorder.Code, recorder.Body.String()), nil
}

// multiRuleCountersScenario shows every rule in one named limiter counts independently.
func multiRuleCountersScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("api", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{
			ratelimit.PerMinute(3).By("user:1"),
			ratelimit.PerMinute(5).By("ip:1"),
		}
	})
	engine := gin.New()
	engine.GET("/api", httpmiddleware.ThrottleFor(limiter, "api"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	for range 2 {
		if recorder := perform(engine, http.MethodGet, "/api"); recorder.Code != http.StatusOK {
			return "", fmt.Errorf("api request status = %d, want 200", recorder.Code)
		}
	}
	ctx := context.Background()
	user, err := limiter.Attempts(ctx, "ratelimit:api:user:1")
	if err != nil {
		return "", fmt.Errorf("read user attempts: %w", err)
	}
	ip, err := limiter.Attempts(ctx, "ratelimit:api:ip:1")
	if err != nil {
		return "", fmt.Errorf("read ip attempts: %w", err)
	}
	return fmt.Sprintf("user=%d ip=%d", user, ip), nil
}

// multiRuleBlockScenario proves any over-limit rule blocks the whole request.
func multiRuleBlockScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("api", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{
			ratelimit.PerMinute(5).By("user:1"),
			ratelimit.PerMinute(1).By("ip:1"),
		}
	})
	engine := gin.New()
	engine.GET("/api", httpmiddleware.ThrottleFor(limiter, "api"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	first := perform(engine, http.MethodGet, "/api").Code
	second := perform(engine, http.MethodGet, "/api").Code
	return fmt.Sprintf("first=%d second=%d", first, second), nil
}
