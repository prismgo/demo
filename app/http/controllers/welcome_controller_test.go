package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/foundation"
	_ "prismgo-demo/config"
)

func TestWelcomeControllerShowReturnsDefaultPayload(t *testing.T) {
	app := foundation.Configure(t.TempDir()).Create()
	defer func() {
		if err := app.CloseContext(context.Background()); err != nil {
			t.Fatalf("close app: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	NewWelcomeController().Show(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["framework"] != "PrismGo" {
		t.Fatalf("framework = %q, want PrismGo", payload["framework"])
	}
	if payload["message"] == "" {
		t.Fatal("message should not be empty")
	}
}
