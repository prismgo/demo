package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"prismgo-demo/app/http/controllers"
	"prismgo-demo/app/providers"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/container"
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
	registerRouteDemoRoutes()
	registerProviderDemoRoutes()
}

// registerProviderDemoRoutes mounts the provider-bound service smoke endpoint.
func registerProviderDemoRoutes() {
	route.Prefix("/api/provider-demo").Group(func() {
		route.Get("/greeting", providerDemoGreeting)
	})
}

// providerDemoGreeting resolves the service bound by AppServiceProvider.Register.
func providerDemoGreeting(c *gin.Context) {
	message, err := container.Make[string](providers.DemoGreetingKey)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "provider service unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// registerRouteDemoRoutes mounts the documented prefix, constraint, naming, binding and resource flow.
func registerRouteDemoRoutes() {
	route.Prefix("/api/route-demo").Name("route-demo.").Group(func() {
		route.Get("/users/{id}", routeDemoUser).WhereNumber("id").Name("users.show")
		route.Bind("post", bindRouteDemoPost)
		route.Get("/posts/{post}", routeDemoShowPost).Name("posts.show")
		route.ApiResource("photos", routeDemoPhotoController{})
		route.Get("/meta/{id}", routeDemoMeta).WhereNumber("id").Name("meta.show")
		route.Redirect("/legacy", "/api/route-demo/posts/7")
	})
}

// routeDemoMeta echoes the injected current-route metadata and its generated named URL.
func routeDemoMeta(c *gin.Context) {
	value, ok := c.Get("route.current")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "current route missing"})
		return
	}
	info, ok := value.(route.RouteInfo)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "current route type mismatch"})
		return
	}
	generated, err := route.URL("route-demo.meta.show", map[string]any{"id": c.Param("id")})
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "url generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name":     info.Name,
		"uri":      info.URI,
		"gin_path": info.GinPath,
		"url":      generated,
	})
}

// routeDemoUser echoes the constrained parameter and the URL generated from its route name.
func routeDemoUser(c *gin.Context) {
	id := c.Param("id")
	generated, err := route.URL("route-demo.users.show", map[string]any{"id": id})
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "url generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "url": generated})
}

// routeDemoPost is the object produced by the route demo post binder.
type routeDemoPost struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// bindRouteDemoPost resolves the post parameter into a demo object, failing on the zero id.
func bindRouteDemoPost(_ *gin.Context, value string) (any, error) {
	if value == "0" {
		return nil, fmt.Errorf("post %s not found", value)
	}
	return &routeDemoPost{ID: value, Title: "post-" + value}, nil
}

// routeDemoShowPost echoes the bound post object.
func routeDemoShowPost(c *gin.Context) {
	value, ok := c.Get("post")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "post not bound"})
		return
	}
	post, ok := value.(*routeDemoPost)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "post type mismatch"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// routeDemoPhotoController implements the API resource actions used by the HTTP smoke test.
type routeDemoPhotoController struct{}

func (routeDemoPhotoController) Index(c *gin.Context) {
	c.String(http.StatusOK, "photos.index")
}

func (routeDemoPhotoController) Store(c *gin.Context) {
	c.String(http.StatusOK, "photos.store")
}

func (routeDemoPhotoController) Show(c *gin.Context) {
	c.String(http.StatusOK, "photos.show:"+c.Param("photo"))
}

func (routeDemoPhotoController) Update(c *gin.Context) {
	c.String(http.StatusOK, "photos.update:"+c.Param("photo"))
}

func (routeDemoPhotoController) Destroy(c *gin.Context) {
	c.String(http.StatusOK, "photos.destroy:"+c.Param("photo"))
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
