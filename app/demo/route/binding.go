package routedemo

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// boundAccount records the object produced by the account parameter binder.
type boundAccount struct {
	ID    string
	Name  string
	Admin bool
}

// bindScenario resolves a path parameter through an explicit binder.
func bindScenario() (string, error) {
	router := route.New()
	router.Bind("user", func(_ *gin.Context, value string) (any, error) {
		return "user:" + value, nil
	})
	router.Get("/users/{user}", boundValueHandler("user"))

	body, err := bodyOf(router, http.MethodGet, "/users/42")
	if err != nil {
		return "", err
	}
	return "bound=" + body, nil
}

// modelScenario resolves a path parameter through the Model alias.
func modelScenario() (string, error) {
	router := route.New()
	router.Model("id", func(_ *gin.Context, value string) (any, error) {
		number, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return number * 2, nil
	})
	router.Get("/double/{id}", boundValueHandler("id"))

	body, err := bodyOf(router, http.MethodGet, "/double/21")
	if err != nil {
		return "", err
	}
	return "model=" + body, nil
}

// bindingContextScenario reads a typed bound object inside the handler.
func bindingContextScenario() (string, error) {
	router := route.New()
	router.Bind("account", func(_ *gin.Context, value string) (any, error) {
		return &boundAccount{ID: value, Name: "acct-" + value, Admin: true}, nil
	})
	router.Get("/accounts/{account}", func(c *gin.Context) {
		value, ok := c.Get("account")
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		account, ok := value.(*boundAccount)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("%s|%s|%t", account.ID, account.Name, account.Admin))
	})

	body, err := bodyOf(router, http.MethodGet, "/accounts/7")
	if err != nil {
		return "", err
	}
	return "value=" + body, nil
}

// missingHandlerScenario runs a custom handler when binding fails.
func missingHandlerScenario() (string, error) {
	router := route.New()
	router.Bind("user", failingBinder("user missing"))
	router.Get("/users/{user}", textHandler("reachable")).Missing(func(c *gin.Context) {
		c.String(http.StatusNotFound, "missing:"+c.Param("user"))
	})

	recorder, err := dispatch(router, http.MethodGet, "/users/404")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("status=%d body=%s", recorder.Code, recorder.Body.String()), nil
}

// defaultMissingScenario falls back to a bare 404 when binding fails without a Missing handler.
func defaultMissingScenario() (string, error) {
	router := route.New()
	router.Bind("user", failingBinder("user missing"))
	router.Get("/users/{user}", textHandler("reachable"))

	recorder, err := dispatch(router, http.MethodGet, "/users/404")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("status=%d body=%s", recorder.Code, recorder.Body.String()), nil
}

// boundValueHandler writes the value stored under param by a parameter binder.
func boundValueHandler(param string) route.HandlerFunc {
	return func(c *gin.Context) {
		value, ok := c.Get(param)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, fmt.Sprint(value))
	}
}

// failingBinder always fails so binding failure handling can be exercised.
func failingBinder(message string) route.Binder {
	return func(*gin.Context, string) (any, error) {
		return nil, errors.New(message)
	}
}
