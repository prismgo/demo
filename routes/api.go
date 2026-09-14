package routes

import (
	"net/http"

	"prismgo-demo/app/http/controllers"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/http/middleware"
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
