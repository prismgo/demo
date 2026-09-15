package routedemo

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
)

// redirectScenario covers default and explicit temporary redirects.
func redirectScenario() (string, error) {
	router := route.New()
	router.Redirect("/old", "/new")
	router.Redirect("/temp", "/target", http.StatusTemporaryRedirect)

	defaulted, err := dispatch(router, http.MethodGet, "/old")
	if err != nil {
		return "", err
	}
	temporary, err := dispatch(router, http.MethodGet, "/temp")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("default=%d:%s temporary=%d:%s",
		defaulted.Code, defaulted.Header().Get("Location"),
		temporary.Code, temporary.Header().Get("Location")), nil
}

// permanentRedirectScenario covers the 301 permanent redirect helper.
func permanentRedirectScenario() (string, error) {
	router := route.New()
	router.PermanentRedirect("/legacy", "/current")

	recorder, err := dispatch(router, http.MethodGet, "/legacy")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("status=%d location=%s", recorder.Code, recorder.Header().Get("Location")), nil
}

// staticScenario serves a file from a temporary directory and rejects a missing file.
func staticScenario() (string, error) {
	root, err := os.MkdirTemp("", "prismgo-route-static-")
	if err != nil {
		return "", fmt.Errorf("create static root: %w", err)
	}
	defer os.RemoveAll(root)
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("static-body"), 0o600); err != nil {
		return "", fmt.Errorf("write static file: %w", err)
	}

	router := route.New()
	router.Static("/assets", root)
	served, err := dispatch(router, http.MethodGet, "/assets/hello.txt")
	if err != nil {
		return "", err
	}
	missing, err := dispatch(router, http.MethodGet, "/assets/missing.txt")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file=%d:%s missing=%d", served.Code, served.Body.String(), missing.Code), nil
}

// fallbackScenario runs the NoRoute handler only for unmatched requests.
func fallbackScenario() (string, error) {
	router := route.New()
	router.Get("/health", textHandler("ok"))
	router.Fallback(func(c *gin.Context) {
		c.String(http.StatusNotFound, "fallback")
	})

	known, err := dispatch(router, http.MethodGet, "/health")
	if err != nil {
		return "", err
	}
	unknown, err := dispatch(router, http.MethodGet, "/missing")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("known=%d:%s unknown=%d:%s",
		known.Code, known.Body.String(), unknown.Code, unknown.Body.String()), nil
}

// domainScenario matches an exact Host header, ignoring the port.
func domainScenario() (string, error) {
	router := route.New()
	router.Domain("api.example.test").Group(func() {
		router.Get("/health", textHandler("health"))
	})

	match, err := dispatchHost(router, http.MethodGet, "/health", "api.example.test")
	if err != nil {
		return "", err
	}
	mismatch, err := dispatchHost(router, http.MethodGet, "/health", "www.example.test")
	if err != nil {
		return "", err
	}
	port, err := dispatchHost(router, http.MethodGet, "/health", "api.example.test:8080")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("match=%d mismatch=%d port=%d", match.Code, mismatch.Code, port.Code), nil
}

// domainPlaceholderScenario matches a single subdomain segment against a placeholder.
func domainPlaceholderScenario() (string, error) {
	router := route.New()
	router.Domain("{tenant}.example.test").Group(func() {
		router.Get("/dashboard", textHandler("dashboard"))
	})

	tenant, err := dispatchHost(router, http.MethodGet, "/dashboard", "acme.example.test")
	if err != nil {
		return "", err
	}
	nested, err := dispatchHost(router, http.MethodGet, "/dashboard", "acme.eu.example.test")
	if err != nil {
		return "", err
	}
	root, err := dispatchHost(router, http.MethodGet, "/dashboard", "example.test")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tenant=%d nested=%d root=%d", tenant.Code, nested.Code, root.Code), nil
}
