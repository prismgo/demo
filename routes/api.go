package routes

import (
	"context"
	"net/http"
	"time"

	"prismgo-demo/app/http/controllers"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/http/middleware"
	"github.com/prismgo/framework/redis"
	"github.com/prismgo/framework/route"
	"github.com/prismgo/framework/session"
)

// Dependencies contains route handlers needed by the route file.
type Dependencies struct {
	WelcomeController *controllers.WelcomeController
}

// Register declares application HTTP routes.
func Register(app Dependencies) {
	route.Get("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if app.WelcomeController != nil {
		route.Get("/api", app.WelcomeController.Show)
	}

	registerSessionDemoRoutes()
	registerRedisDemoRoutes()
}

// registerRedisDemoRoutes mounts the documented Facade client flow for smoke testing.
func registerRedisDemoRoutes() {
	route.Prefix("/api/redis-demo").Group(func() {
		route.Get("/counter", redisDemoCounter)
	})
}

// redisDemoCounter increments a demo key through the Facade client and returns the value.
func redisDemoCounter(c *gin.Context) {
	client, err := redis.Client()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "redis unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	key := "prismgo_demo_http:counter"
	value, err := client.Incr(ctx, key).Result()
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "redis command failed"})
		return
	}
	if err := client.Expire(ctx, key, time.Minute).Err(); err != nil {
		_ = c.Error(err)
	}
	c.JSON(http.StatusOK, gin.H{"counter": value})
}

// registerSessionDemoRoutes mounts the documented StartSession flow for smoke testing.
func registerSessionDemoRoutes() {
	route.Prefix("/api/session-demo").
		Middleware(middleware.StartSession()).
		Group(func() {
			route.Get("/profile", sessionDemoProfile)
			route.Post("/profile", sessionDemoUpdateProfile)
		})
}

// sessionDemoProfile reads persisted session values; the flash notice disappears after this read.
func sessionDemoProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id": session.Get(c, "user_id", int64(0)),
		"notice":  session.Pull(c, "notice", ""),
	})
}

// sessionDemoUpdateProfile stores a user ID and queues a flash notice for the next request.
func sessionDemoUpdateProfile(c *gin.Context) {
	if err := session.Put(c, "user_id", int64(1001)); err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if err := session.Flash(c, "notice", "saved"); err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}
