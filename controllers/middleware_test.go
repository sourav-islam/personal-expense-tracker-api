package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"expense-tracker-api/models"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// setupMiddlewareTestEnv creates a temp users CSV with one seeded user.
func setupMiddlewareTestEnv(t *testing.T) func() {
	t.Helper()
	tmpPath := t.TempDir() + "/users_test.csv"
	beego.AppConfig.Set("users_csv_path", tmpPath)
	_ = models.CreateUser(&models.User{
		Name: "Alice", Email: "alice@example.com", Password: "pass123",
	})
	return func() { os.Remove(tmpPath) }
}

// newBaseController builds a BaseController wired to a fake HTTP request/response.
func newBaseController(method, path, userIDHeader string) *BaseController {
	req := httptest.NewRequest(method, path, nil)
	if userIDHeader != "" {
		req.Header.Set("X-User-ID", userIDHeader)
	}
	rw := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(rw, req)

	c := &BaseController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	return c
}

func TestAuthMiddleware(t *testing.T) {
	cleanup := setupMiddlewareTestEnv(t)
	defer cleanup()

	tests := []struct {
		name         string
		userIDHeader string
		wantUserID   int
		wantStatus   int
	}{
		{name: "valid user ID", userIDHeader: "1", wantUserID: 1, wantStatus: 0},
		{name: "missing header", userIDHeader: "", wantUserID: 0, wantStatus: 401},
		{name: "non-numeric ID", userIDHeader: "abc", wantUserID: 0, wantStatus: 401},
		{name: "zero ID", userIDHeader: "0", wantUserID: 0, wantStatus: 401},
		{name: "negative ID", userIDHeader: "-1", wantUserID: 0, wantStatus: 401},
		{name: "non-existing user", userIDHeader: "999", wantUserID: 0, wantStatus: 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newBaseController(http.MethodGet, "/api/v1/expenses", tt.userIDHeader)
			got := AuthMiddleware(c)
			if got != tt.wantUserID {
				t.Errorf("AuthMiddleware() = %d, want %d", got, tt.wantUserID)
			}
		})
	}
}
