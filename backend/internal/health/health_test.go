package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestHealthReturnsOKWithoutDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checker := New(nil, nil, zap.NewNop())

	r := gin.New()
	r.GET("/health", checker.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Checks["server"] != "healthy" {
		t.Fatalf("server check = %q", body.Checks["server"])
	}
	if body.Checks["database"] != "not_configured" {
		t.Fatalf("database check = %q", body.Checks["database"])
	}
	if body.Checks["redis"] != "not_configured" {
		t.Fatalf("redis check = %q", body.Checks["redis"])
	}
}

func TestLiveReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checker := New(nil, nil, zap.NewNop())

	r := gin.New()
	r.GET("/live", checker.Live)

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}
