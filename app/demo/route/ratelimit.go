package routedemo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/ratelimit"
	"github.com/prismgo/framework/route"
)

// rateLimiterScenario registers a named limiter and reads it back from the running application.
func rateLimiterScenario() (string, error) {
	route.RateLimiter("demo-route-limiter", func(*gin.Context) []route.Limit {
		return []route.Limit{route.PerMinute(5).By(func(*gin.Context) string { return "user:42" })}
	})

	router := route.New()
	router.Post("/login", route.Throttle("demo-route-limiter"), textHandler("ok"))
	status, err := statusOf(router, http.MethodPost, "/login")
	if err != nil {
		return "", err
	}

	resolver := ratelimit.Limiter("demo-route-limiter")
	if resolver == nil {
		return "", fmt.Errorf("named limiter demo-route-limiter is not registered")
	}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
	limits := resolver(ctx)
	if len(limits) != 1 {
		return "", fmt.Errorf("named limiter limits = %d, want 1", len(limits))
	}
	return fmt.Sprintf("max=%d every=%s key=%s status=%d",
		limits[0].MaxAttempts, limits[0].Decay, limits[0].Key, status), nil
}

// limitScenario inspects the Limit window and key builder without a running application.
func limitScenario() (string, error) {
	limit := route.PerMinute(3).By(func(*gin.Context) string { return "tenant:9" })
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	return fmt.Sprintf("max=%d every=%s key=%s", limit.Max, limit.Every, limit.KeyFunc(ctx)), nil
}

// throttleRouteScenario prevents route-level requests once the route limiter is exhausted.
func throttleRouteScenario() (string, error) {
	route.RateLimiter("demo-throttle-route", func(*gin.Context) []route.Limit {
		return []route.Limit{route.PerMinute(2)}
	})

	router := route.New()
	router.Get("/ping", route.Throttle("demo-throttle-route"), textHandler("pong"))

	statuses := make([]string, 0, 3)
	for range 3 {
		status, err := statusOf(router, http.MethodGet, "/ping")
		if err != nil {
			return "", err
		}
		statuses = append(statuses, fmt.Sprint(status))
	}
	return "statuses=" + strings.Join(statuses, ","), nil
}

// throttleGroupScenario shares one limiter window across every route in a group.
func throttleGroupScenario() (string, error) {
	route.RateLimiter("demo-throttle-group", func(*gin.Context) []route.Limit {
		return []route.Limit{route.PerMinute(2)}
	})

	router := route.New()
	router.Middleware(route.Throttle("demo-throttle-group")).Group(func() {
		router.Get("/a", textHandler("a"))
		router.Get("/b", textHandler("b"))
	})

	first, err := statusOf(router, http.MethodGet, "/a")
	if err != nil {
		return "", err
	}
	second, err := statusOf(router, http.MethodGet, "/a")
	if err != nil {
		return "", err
	}
	third, err := statusOf(router, http.MethodGet, "/b")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("a1=%d a2=%d b1=%d", first, second, third), nil
}

// throttleUnknownScenario lets requests pass when the named limiter is not registered.
func throttleUnknownScenario() (string, error) {
	router := route.New()
	router.Get("/open", route.Throttle("demo-missing-limiter"), textHandler("open"))

	recorder, err := dispatch(router, http.MethodGet, "/open")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("status=%d body=%s", recorder.Code, recorder.Body.String()), nil
}

// throttleOverLimitScenario exposes the 429 status and over-limit headers.
func throttleOverLimitScenario() (string, error) {
	route.RateLimiter("demo-throttle-tight", func(*gin.Context) []route.Limit {
		return []route.Limit{route.PerMinute(1)}
	})

	router := route.New()
	router.Get("/tight", route.Throttle("demo-throttle-tight"), textHandler("tight"))

	first, err := dispatch(router, http.MethodGet, "/tight")
	if err != nil {
		return "", err
	}
	second, err := dispatch(router, http.MethodGet, "/tight")
	if err != nil {
		return "", err
	}
	retry := second.Header().Get("Retry-After")
	return fmt.Sprintf("first=%d second=%d remaining=%s retry-positive=%t",
		first.Code, second.Code, second.Header().Get("X-RateLimit-Remaining"), retry != "" && retry != "0"), nil
}
