package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
	"prismgo-demo/app/http/controllers"
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
}
