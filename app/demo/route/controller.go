package routedemo

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// actionController exposes Gin actions resolved by method name.
type actionController struct{}

func (actionController) Index(c *gin.Context) { c.String(http.StatusOK, "index") }
func (actionController) Show(c *gin.Context)  { c.String(http.StatusOK, "show:"+c.Param("id")) }

// invalidController carries a wrong-signature action so controller validation can be exercised.
type invalidController struct{}

func (invalidController) Bad(*gin.Context) int { return 0 }

// controllerActionScenario registers controller handlers by action name.
func controllerActionScenario() (string, error) {
	router := route.New()
	users := router.Controller(actionController{}).Prefix("/users")
	users.Action(http.MethodGet, "/", "Index")
	users.Action(http.MethodGet, "/{id}", "Show")

	index, err := bodyOf(router, http.MethodGet, "/users")
	if err != nil {
		return "", err
	}
	show, err := bodyOf(router, http.MethodGet, "/users/7")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("index=%s show=%s", index, show), nil
}

// controllerValidationScenario captures the startup panics for invalid controller actions.
func controllerValidationScenario() (string, error) {
	router := route.New()
	missing := panicMessage(func() {
		router.Controller(actionController{}).Action(http.MethodGet, "/x", "Missing")
	})
	signature := panicMessage(func() {
		router.Controller(invalidController{}).Action(http.MethodGet, "/y", "Bad")
	})
	unconfigured := panicMessage(func() {
		router.Controller(nil).Action(http.MethodGet, "/z", "Index")
	})
	return fmt.Sprintf("missing=%q signature=%q unconfigured=%q", missing, signature, unconfigured), nil
}

// resourceControllerContractScenario asserts the documented resource controller interface.
func resourceControllerContractScenario() (string, error) {
	var _ route.ResourceController = demoResourceController{}
	return interfaceSummary(reflect.TypeOf((*route.ResourceController)(nil)).Elem()), nil
}

// createControllerContractScenario asserts the documented optional create interface.
func createControllerContractScenario() (string, error) {
	var _ route.CreateController = demoResourceController{}
	return interfaceSummary(reflect.TypeOf((*route.CreateController)(nil)).Elem()), nil
}

// editControllerContractScenario asserts the documented optional edit interface.
func editControllerContractScenario() (string, error) {
	var _ route.EditController = demoResourceController{}
	return interfaceSummary(reflect.TypeOf((*route.EditController)(nil)).Elem()), nil
}

// panicMessage runs fn and returns the recovered panic message, or an empty string.
func panicMessage(fn func()) (message string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			message = fmt.Sprint(recovered)
		}
	}()
	fn()
	return ""
}

// interfaceSummary renders an interface name with its alphabetically ordered methods.
func interfaceSummary(contract reflect.Type) string {
	methods := make([]string, 0, contract.NumMethod())
	for index := 0; index < contract.NumMethod(); index++ {
		methods = append(methods, contract.Method(index).Name)
	}
	return fmt.Sprintf("interface=%s methods=%s", contract.Name(), strings.Join(methods, ","))
}
