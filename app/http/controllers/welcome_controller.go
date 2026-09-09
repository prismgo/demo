// Package controllers contains the demo application's HTTP controllers.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prismgo-demo/config"
)

// WelcomeController handles the default API landing endpoint.
type WelcomeController struct{}

// NewWelcomeController creates the default welcome controller.
func NewWelcomeController() *WelcomeController {
	return &WelcomeController{}
}

// Show returns a small JSON payload proving that the application is running.
func (c *WelcomeController) Show(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"framework": "PrismGo",
		"name":      config.GetString("app.name", "PrismGo"),
		"message":   "Welcome to PrismGo.",
	})
}
