package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"prismgo-demo/app/http/controllers"
	_ "prismgo-demo/config"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/foundation"
	"github.com/prismgo/framework/route"
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

	assertSessionDemoFlow(t, engine)
	assertRouteRegistered(t, engine, http.MethodGet, "/api/redis-demo/counter")
	assertRouteRegistered(t, engine, http.MethodGet, "/api/route-demo/meta/:id")
}

// assertRouteRegistered verifies a mounted route path is present.
func assertRouteRegistered(t *testing.T, engine *gin.Engine, method string, path string) {
	t.Helper()
	for _, registered := range engine.Routes() {
		if registered.Method == method && registered.Path == path {
			return
		}
	}
	t.Fatalf("route %s %s is not registered; routes = %#v", method, path, engine.Routes())
}

// assertSessionDemoFlow drives the /api/session-demo group across three requests.
func assertSessionDemoFlow(t *testing.T, engine *gin.Engine) {
	t.Helper()

	update := performSessionRequest(t, engine, http.MethodPost, "/api/session-demo/profile", nil)
	if update.Code != http.StatusNoContent {
		t.Fatalf("session update status = %d, want %d", update.Code, http.StatusNoContent)
	}
	idCookie := sessionCookie(t, update)

	first := performSessionRequest(t, engine, http.MethodGet, "/api/session-demo/profile", []*http.Cookie{idCookie})
	var flashBody map[string]any
	if err := json.Unmarshal(first.Body.Bytes(), &flashBody); err != nil {
		t.Fatalf("decode flash payload: %v\nbody = %q", err, first.Body.String())
	}
	if flashBody["user_id"].(float64) != 1001 || flashBody["notice"] != "saved" {
		t.Fatalf("first profile body = %v, want user_id 1001 and notice saved", flashBody)
	}
	if again := sessionCookie(t, first); again.Value != idCookie.Value {
		t.Fatalf("session id cookie changed between requests: got %q, want %q", again.Value, idCookie.Value)
	}

	second := performSessionRequest(t, engine, http.MethodGet, "/api/session-demo/profile", []*http.Cookie{idCookie})
	var clearedBody map[string]any
	if err := json.Unmarshal(second.Body.Bytes(), &clearedBody); err != nil {
		t.Fatalf("decode cleared payload: %v\nbody = %q", err, second.Body.String())
	}
	if clearedBody["user_id"].(float64) != 1001 || clearedBody["notice"] != "" {
		t.Fatalf("second profile body = %v, want user_id 1001 and empty notice", clearedBody)
	}
}

// performSessionRequest executes one request carrying the provided cookies.
func performSessionRequest(t *testing.T, engine *gin.Engine, method string, path string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	engine.ServeHTTP(recorder, request)
	return recorder
}

// sessionCookie extracts the prismgo_session cookie from a response.
func sessionCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "prismgo_session" {
			return cookie
		}
	}
	t.Fatalf("response has no prismgo_session cookie; Set-Cookie headers = %q", recorder.Result().Header.Values("Set-Cookie"))
	return nil
}

func performRequest(engine *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, nil)
	engine.ServeHTTP(recorder, request)
	return recorder
}
