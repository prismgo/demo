package ratelimitdemo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	httpmiddleware "github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/ratelimit"
)

// byScenario shows that each By dimension key keeps an independent counter.
func byScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("scope", func(c *gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{ratelimit.PerMinute(1).By(strings.TrimSpace(c.Query("tenant")))}
	})
	engine := gin.New()
	engine.GET("/scope", httpmiddleware.ThrottleFor(limiter, "scope"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	firstA := perform(engine, http.MethodGet, "/scope?tenant=a").Code
	secondA := perform(engine, http.MethodGet, "/scope?tenant=a").Code
	firstB := perform(engine, http.MethodGet, "/scope?tenant=b").Code
	return fmt.Sprintf("a=%d,%d b=%d", firstA, secondA, firstB), nil
}

// fallbackKeyScenario shows a duplicate rule key switching to its fallback key.
func fallbackKeyScenario() (string, error) {
	manager, limiter, err := newTestLimiter()
	if err != nil {
		return "", err
	}
	defer func() { _ = manager.Close() }()

	limiter.For("mixed", func(*gin.Context) []ratelimit.Limit {
		return []ratelimit.Limit{
			ratelimit.PerMinute(1).By("same"),
			ratelimit.PerMinute(2).By("same").FallbackKey("same:fallback"),
		}
	})
	engine := gin.New()
	engine.GET("/mixed", httpmiddleware.ThrottleFor(limiter, "mixed"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	first := perform(engine, http.MethodGet, "/mixed").Code

	ctx := context.Background()
	primary, err := limiter.Attempts(ctx, "ratelimit:mixed:same")
	if err != nil {
		return "", fmt.Errorf("read primary attempts: %w", err)
	}
	fallback, err := limiter.Attempts(ctx, "ratelimit:mixed:same:fallback")
	if err != nil {
		return "", fmt.Errorf("read fallback attempts: %w", err)
	}
	second := perform(engine, http.MethodGet, "/mixed").Code
	return fmt.Sprintf("primary=%d fallback=%d first=%d second=%d", primary, fallback, first, second), nil
}
