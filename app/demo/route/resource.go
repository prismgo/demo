package routedemo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// demoResourceController implements every resource action used by the route demo.
type demoResourceController struct{}

func (demoResourceController) Index(c *gin.Context)   { c.String(http.StatusOK, "index") }
func (demoResourceController) Store(c *gin.Context)   { c.String(http.StatusOK, "store") }
func (demoResourceController) Show(c *gin.Context)    { c.String(http.StatusOK, "show") }
func (demoResourceController) Update(c *gin.Context)  { c.String(http.StatusOK, "update") }
func (demoResourceController) Destroy(c *gin.Context) { c.String(http.StatusOK, "destroy") }
func (demoResourceController) Create(c *gin.Context)  { c.String(http.StatusOK, "create") }
func (demoResourceController) Edit(c *gin.Context)    { c.String(http.StatusOK, "edit") }

// apiResourceScenario registers the five default API resource actions.
func apiResourceScenario() (string, error) {
	router := route.New()
	router.ApiResource("photos", demoResourceController{})
	return resourceResult(router, "photos"), nil
}

// resourceScenario registers the full resource including create and edit actions.
func resourceScenario() (string, error) {
	router := route.New()
	router.Resource("photos", demoResourceController{})
	return resourceResult(router, "photos"), nil
}

// resourceCreateScenario registers only the optional create action.
func resourceCreateScenario() (string, error) {
	router := route.New()
	router.Resource("photos", demoResourceController{}, route.Only("create"))
	create, err := statusOf(router, http.MethodGet, "/photos/create")
	if err != nil {
		return "", err
	}
	list, err := statusOf(router, http.MethodGet, "/photos")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("routes=%d create=%d list=%d", len(resourceRoutes(router.List(), "photos")), create, list), nil
}

// resourceEditScenario registers only the optional edit action.
func resourceEditScenario() (string, error) {
	router := route.New()
	router.Resource("photos", demoResourceController{}, route.Only("edit"))
	edit, err := statusOf(router, http.MethodGet, "/photos/9/edit")
	if err != nil {
		return "", err
	}
	list, err := statusOf(router, http.MethodGet, "/photos")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("routes=%d edit=%d list=%d", len(resourceRoutes(router.List(), "photos")), edit, list), nil
}

// resourceOnlyScenario keeps a subset of API resource actions.
func resourceOnlyScenario() (string, error) {
	router := route.New()
	router.ApiResource("photos", demoResourceController{}, route.Only("index", "show"))
	index, err := statusOf(router, http.MethodGet, "/photos")
	if err != nil {
		return "", err
	}
	show, err := statusOf(router, http.MethodGet, "/photos/9")
	if err != nil {
		return "", err
	}
	store, err := statusOf(router, http.MethodPost, "/photos")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("routes=%d index=%d show=%d store=%d", len(resourceRoutes(router.List(), "photos")), index, show, store), nil
}

// resourceExceptScenario drops one API resource action.
func resourceExceptScenario() (string, error) {
	router := route.New()
	router.ApiResource("photos", demoResourceController{}, route.Except("destroy"))
	destroy, err := statusOf(router, http.MethodDelete, "/photos/9")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("routes=%d destroy=%d", len(resourceRoutes(router.List(), "photos")), destroy), nil
}

// resourceNamesScenario overrides individual resource route names.
func resourceNamesScenario() (string, error) {
	router := route.New()
	router.ApiResource("photos", demoResourceController{}, route.Names(map[string]string{
		"index": "photos.list",
		"show":  "photos.detail",
	}))
	index, err := router.URL("photos.list", nil)
	if err != nil {
		return "", err
	}
	show, err := router.URL("photos.detail", map[string]any{"photo": 9})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("index=%s show=%s", index, show), nil
}

// resourceParametersScenario overrides the resource member parameter name.
func resourceParametersScenario() (string, error) {
	router := route.New()
	router.ApiResource("photos", demoResourceController{}, route.Parameters(map[string]string{
		"photos": "photo_id",
	}))
	uri := ""
	for _, info := range resourceRoutes(router.List(), "photos") {
		if info.Name == "photos.show" {
			uri = info.URI
		}
	}
	generated, err := router.URL("photos.show", map[string]any{"photo_id": 9})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("uri=%s url=%s", uri, generated), nil
}

// apiResourcesScenario registers several API resources in one call.
func apiResourcesScenario() (string, error) {
	router := route.New()
	router.ApiResources(map[string]route.ResourceController{
		"photos": demoResourceController{},
		"posts":  demoResourceController{},
	})
	photos, err := router.URL("photos.show", map[string]any{"photo": 8})
	if err != nil {
		return "", err
	}
	posts, err := router.URL("posts.show", map[string]any{"post": 9})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("routes=%d photos=%s posts=%s", len(router.List()), photos, posts), nil
}

// nestedResourceScenario registers a dotted resource name as a nested path.
func nestedResourceScenario() (string, error) {
	router := route.New()
	router.ApiResource("users.photos", demoResourceController{})
	index, err := router.URL("users.photos.index", nil)
	if err != nil {
		return "", err
	}
	show, err := router.URL("users.photos.show", map[string]any{"photo": 9})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("routes=%d index=%s show=%s", len(router.List()), index, show), nil
}

// resourceResult renders the routes registered for one resource prefix.
func resourceResult(router *route.Router, prefix string) string {
	routes := resourceRoutes(router.List(), prefix)
	return fmt.Sprintf("routes=%d %s", len(routes), resourceSummary(routes))
}

// resourceRoutes keeps only the snapshots whose name belongs to the resource prefix.
func resourceRoutes(infos []route.RouteInfo, prefix string) []route.RouteInfo {
	routes := make([]route.RouteInfo, 0, len(infos))
	for _, info := range infos {
		if strings.HasPrefix(info.Name, prefix+".") {
			routes = append(routes, info)
		}
	}
	return routes
}

// resourceSummary renders the method, URI and name of each resource route.
func resourceSummary(infos []route.RouteInfo) string {
	parts := make([]string, 0, len(infos))
	for _, info := range infos {
		parts = append(parts, fmt.Sprintf("%s %s=%s", strings.Join(info.Methods, ","), info.URI, info.Name))
	}
	return strings.Join(parts, ";")
}
