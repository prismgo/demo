package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/route"
	"prismgo-demo/app/http/controllers"
	_ "prismgo-demo/config"
)

func TestRegisterAddsHealthAndWelcomeRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := foundation.Configure(t.TempDir()).Create()
	defer func() {
		if err := app.CloseContext(context.Background()); err != nil {
			t.Fatalf("close app: %v", err)
		}
	}()
	if err := app.Boot(); err != nil {
		t.Fatalf("Boot() error = %v", err)
	}

	Register(Dependencies{WelcomeController: controllers.NewWelcomeController()})

	engine := gin.New()
	if err := route.Mount(engine); err != nil {
		t.Fatalf("mount routes: %v", err)
	}

	health := performRequest(engine, http.MethodGet, "/api/health")
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", health.Code, http.StatusOK)
	}

	welcome := performRequest(engine, http.MethodGet, "/api")
	if welcome.Code != http.StatusOK {
		t.Fatalf("welcome status = %d, want %d", welcome.Code, http.StatusOK)
	}

	var payload map[string]string
	if err := json.Unmarshal(welcome.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode welcome payload: %v", err)
	}
	if payload["framework"] != "PrismGo" {
		t.Fatalf("framework = %q, want PrismGo", payload["framework"])
	}
}

func performRequest(engine *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, nil)
	engine.ServeHTTP(recorder, request)
	return recorder
}
