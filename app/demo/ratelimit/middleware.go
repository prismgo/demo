package ratelimitdemo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	httpmiddleware "github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/ratelimit"
	"github.com/prismgo/framework/route"
)

// afterCountScenario counts matching responses through the After callback.
func afterCountScenario() (string, error) {
	ratelimit.For("demo-after-count", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(2).By("after-count").After(func(c *gin.Context) bool {
			return c.Writer.Status() >= http.StatusInternalServerError
		})}
	})
	engine := gin.New()
	engine.GET("/boom", httpmiddleware.Throttle("demo-after-count"), func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "boom")
	})

	first := perform(engine, http.MethodGet, "/boom").Code
	second := perform(engine, http.MethodGet, "/boom").Code
	third := perform(engine, http.MethodGet, "/boom").Code
	attempts, err := ratelimit.Attempts(context.Background(), "ratelimit:demo-after-count:after-count")
	if err != nil {
		return "", fmt.Errorf("read after-count attempts: %w", err)
	}
	return fmt.Sprintf("first=%d second=%d third=%d attempts=%d", first, second, third, attempts), nil
}

// afterSkipScenario leaves non-matching responses uncounted by the After callback.
func afterSkipScenario() (string, error) {
	ratelimit.For("demo-after-skip", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(2).By("after-skip").After(func(c *gin.Context) bool {
			return c.Writer.Status() >= http.StatusInternalServerError
		})}
	})
	engine := gin.New()
	engine.GET("/ok", httpmiddleware.Throttle("demo-after-skip"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	engine.GET("/boom", httpmiddleware.Throttle("demo-after-skip"), func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "boom")
	})

	ok := perform(engine, http.MethodGet, "/ok").Code
	afterOK, err := ratelimit.Attempts(context.Background(), "ratelimit:demo-after-skip:after-skip")
	if err != nil {
		return "", fmt.Errorf("read after-skip attempts after success: %w", err)
	}
	serverError := perform(engine, http.MethodGet, "/boom").Code
	afterError, err := ratelimit.Attempts(context.Background(), "ratelimit:demo-after-skip:after-skip")
	if err != nil {
		return "", fmt.Errorf("read after-skip attempts after server error: %w", err)
	}
	return fmt.Sprintf("ok=%d ok-attempts=%d server-error=%d error-attempts=%d",
		ok, afterOK, serverError, afterError), nil
}

// afterHeadersScenario shows After rules expose pre-count success headers.
func afterHeadersScenario() (string, error) {
	ratelimit.For("demo-after-headers", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(3).By("after-headers").After(func(*gin.Context) bool {
			return true
		})}
	})
	engine := gin.New()
	engine.GET("/always", httpmiddleware.Throttle("demo-after-headers"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := perform(engine, http.MethodGet, "/always")
	attempts, err := ratelimit.Attempts(context.Background(), "ratelimit:demo-after-headers:after-headers")
	if err != nil {
		return "", fmt.Errorf("read after-headers attempts: %w", err)
	}
	return fmt.Sprintf("status=%d limit=%s remaining=%s attempts=%d",
		recorder.Code,
		recorder.Header().Get("X-RateLimit-Limit"),
		recorder.Header().Get("X-RateLimit-Remaining"),
		attempts), nil
}

// customResponseScenario serves a custom over-limit response from the Response callback.
func customResponseScenario() (string, error) {
	ratelimit.For("demo-custom-response", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(1).By("custom-response").Response(func(c *gin.Context, result ratelimit.Result) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"custom": true, "max": result.MaxAttempts})
		})}
	})
	engine := gin.New()
	engine.GET("/custom", httpmiddleware.Throttle("demo-custom-response"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	first := perform(engine, http.MethodGet, "/custom")
	second := perform(engine, http.MethodGet, "/custom")
	return fmt.Sprintf("first=%d second=%d body=%s", first.Code, second.Code, second.Body.String()), nil
}

// throttleForScenario mounts two explicit limiters that share a limiter name.
func throttleForScenario() (string, error) {
	managerA, limiterA, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = managerA.Close() }()
	managerB, limiterB, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = managerB.Close() }()

	limiterA.For("shared", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(1).By("shared")}
	})
	limiterB.For("shared", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(2).By("shared")}
	})

	engineA := gin.New()
	engineA.GET("/a", httpmiddleware.ThrottleFor(limiterA, "shared"), textOK)
	engineB := gin.New()
	engineB.GET("/b", httpmiddleware.ThrottleFor(limiterB, "shared"), textOK)

	firstA := perform(engineA, http.MethodGet, "/a").Code
	secondA := perform(engineA, http.MethodGet, "/a").Code
	firstB := perform(engineB, http.MethodGet, "/b").Code
	secondB := perform(engineB, http.MethodGet, "/b").Code
	return fmt.Sprintf("limiter-a=%d,%d limiter-b=%d,%d", firstA, secondA, firstB, secondB), nil
}

// disabledRuleScenario lets requests pass when every rule is disabled.
func disabledRuleScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("disabled", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{
			ratelimit.None(),
			ratelimit.Every(0, 5),
			ratelimit.Every(time.Minute, 0),
		}
	})
	engine := gin.New()
	engine.GET("/disabled", httpmiddleware.ThrottleFor(limiter, "disabled"), textOK)

	first := perform(engine, http.MethodGet, "/disabled")
	second := perform(engine, http.MethodGet, "/disabled")
	return fmt.Sprintf("first=%d second=%d has-limit-header=%t",
		first.Code, second.Code, first.Header().Get("X-RateLimit-Limit") != ""), nil
}

// defaultResponseScenario renders the framework's default over-limit response.
func defaultResponseScenario() (string, error) {
	ratelimit.For("demo-default-response", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(1).By("default-response")}
	})
	engine := gin.New()
	engine.GET("/limited", httpmiddleware.Throttle("demo-default-response"), textOK)

	first := perform(engine, http.MethodGet, "/limited")
	second := perform(engine, http.MethodGet, "/limited")
	return fmt.Sprintf("first=%d second=%d body=%s", first.Code, second.Code, second.Body.String()), nil
}

// successHeadersScenario exposes the rate-limit headers of an allowed request.
func successHeadersScenario() (string, error) {
	ratelimit.For("demo-success-headers", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(3).By("success-headers")}
	})
	engine := gin.New()
	engine.GET("/ok", httpmiddleware.Throttle("demo-success-headers"), textOK)

	recorder := perform(engine, http.MethodGet, "/ok")
	return fmt.Sprintf("status=%d limit=%s remaining=%s",
		recorder.Code,
		recorder.Header().Get("X-RateLimit-Limit"),
		recorder.Header().Get("X-RateLimit-Remaining")), nil
}

// overLimitHeadersScenario exposes the headers of a blocked request.
func overLimitHeadersScenario() (string, error) {
	ratelimit.For("demo-over-limit-headers", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(1).By("over-limit-headers")}
	})
	engine := gin.New()
	engine.GET("/limit", httpmiddleware.Throttle("demo-over-limit-headers"), textOK)

	_ = perform(engine, http.MethodGet, "/limit")
	blocked := perform(engine, http.MethodGet, "/limit")
	return fmt.Sprintf("status=%d limit=%s remaining=%s retry-positive=%t reset-positive=%t",
		blocked.Code,
		blocked.Header().Get("X-RateLimit-Limit"),
		blocked.Header().Get("X-RateLimit-Remaining"),
		positiveHeader(blocked, "Retry-After"),
		positiveHeader(blocked, "X-RateLimit-Reset")), nil
}

// tightestHeadersScenario reports the header of the rule with least remaining.
func tightestHeadersScenario() (string, error) {
	ratelimit.For("demo-tightest-headers", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{
			ratelimit.PerMinute(5).By("wide"),
			ratelimit.PerMinute(2).By("tight"),
		}
	})
	engine := gin.New()
	engine.GET("/mixed", httpmiddleware.Throttle("demo-tightest-headers"), textOK)

	recorder := perform(engine, http.MethodGet, "/mixed")
	return fmt.Sprintf("limit=%s remaining=%s",
		recorder.Header().Get("X-RateLimit-Limit"),
		recorder.Header().Get("X-RateLimit-Remaining")), nil
}

// routeGroupScenario shares one limiter window across a Gin route group.
func routeGroupScenario() (string, error) {
	ratelimit.For("demo-route-group", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(2).By("route-group")}
	})
	engine := gin.New()
	group := engine.Group("/api", httpmiddleware.Throttle("demo-route-group"))
	group.GET("/a", textOK)
	group.GET("/b", textOK)

	firstA := perform(engine, http.MethodGet, "/api/a").Code
	secondA := perform(engine, http.MethodGet, "/api/a").Code
	firstB := perform(engine, http.MethodGet, "/api/b").Code
	return fmt.Sprintf("a1=%d a2=%d b1=%d", firstA, secondA, firstB), nil
}

// routeCompatibilityScenario mounts a limiter through the route compatibility layer.
func routeCompatibilityScenario() (string, error) {
	route.RateLimiter("demo-route-compat", func(*gin.Context) []route.Limit {
		return []route.Limit{route.PerMinute(2).By(func(*gin.Context) string { return "user:42" })}
	})
	router := route.New()
	router.Get("/compat", route.Throttle("demo-route-compat"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	statuses := make([]string, 0, 3)
	for range 3 {
		recorder, err := mountAndPerform(router, http.MethodGet, "/compat")
		if err != nil {
			return "", err
		}
		statuses = append(statuses, strconv.Itoa(recorder.Code))
	}
	return fmt.Sprintf("max=2 statuses=%s", strings.Join(statuses, ",")), nil
}

// mountAndPerform mounts a router on a fresh engine and performs one request.
func mountAndPerform(router *route.Router, method, target string) (*httptest.ResponseRecorder, error) {
	engine := gin.New()
	if err := router.Mount(engine); err != nil {
		return nil, fmt.Errorf("mount router: %w", err)
	}
	return perform(engine, method, target), nil
}

// positiveHeader reports whether a response header holds a positive integer.
func positiveHeader(recorder *httptest.ResponseRecorder, name string) bool {
	value, err := strconv.Atoi(strings.TrimSpace(recorder.Header().Get(name)))
	return err == nil && value > 0
}

// textOK responds with a fixed success body.
func textOK(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// throttleScenario mounts the global named limiter as Gin middleware.
func throttleScenario() (string, error) {
	ratelimit.For("demo-global-throttle", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(1).By("global-throttle")}
	})
	engine := gin.New()
	engine.GET("/ping", httpmiddleware.Throttle("demo-global-throttle"), func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	first := perform(engine, http.MethodGet, "/ping")
	second := perform(engine, http.MethodGet, "/ping")
	return fmt.Sprintf("first=%d remaining=%s retry-after=%t second=%d",
		first.Code,
		first.Header().Get("X-RateLimit-Remaining"),
		second.Header().Get("Retry-After") != "",
		second.Code), nil
}
