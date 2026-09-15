package routedemo

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// bestPracticesScenario composes the recommended prefix, name, middleware and constraint pattern.
func bestPracticesScenario() (string, error) {
	router := route.New()
	router.Prefix("/api/v1").Name("api.").Middleware(func(c *gin.Context) { c.Next() }).Group(func() {
		router.Get("/prismgos/{id}", textHandler("prismgos")).WhereNumber("id").Name("prismgos.show")
	})

	name := ""
	for _, info := range router.List() {
		name = info.Name
	}
	generated, err := router.URL("api.prismgos.show", map[string]any{"id": 7})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("url=%s routes=%d name=%s", generated, len(router.List()), name), nil
}
