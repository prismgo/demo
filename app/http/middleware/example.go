package middleware

import "github.com/gin-gonic/gin"

// Example adds a simple marker header to responses passing through it.
func Example() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Prismgo", "starter")
		c.Next()
	}
}
