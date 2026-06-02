package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"expense-tracker-api/models"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// setupAuthTestEnv creates a temp CSV and seeds one existing user.
func setupAuthTestEnv(t *testing.T) func() {
	t.Helper()
	tmpPath := t.TempDir() + "/users_test.csv"
	beego.AppConfig.Set("users_csv_path", tmpPath)
	_ = models.CreateUser(&models.User{
		Name: "Existing", Email: "existing@example.com", Password: "pass123",
	})
	return func() { os.Remove(tmpPath) }
}

// newAuthController wires an AuthController to a fake HTTP request with a JSON body.
func newAuthController(method, path string, body []byte) *AuthController {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(rw, req)
	ctx.Input.RequestBody = body

	c := &AuthController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	return c
}

// responseStatus reads the HTTP status code written by SendError/SendSuccess.
func responseStatus(c *AuthController) int {
	return c.Ctx.ResponseWriter.Status
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestRegister(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid registration",
			body:       `{"name":"John Doe","email":"john@example.com","password":"secret123"}`,
			wantStatus: 201,
		},
		{
			name:       "duplicate email",
			body:       `{"name":"Other","email":"existing@example.com","password":"secret123"}`,
			wantStatus: 409,
		},
		{
			name:       "missing name",
			body:       `{"email":"john2@example.com","password":"secret123"}`,
			wantStatus: 400,
		},
		{
			name:       "missing email",
			body:       `{"name":"John","password":"secret123"}`,
			wantStatus: 400,
		},
		{
			name:       "invalid email format",
			body:       `{"name":"John","email":"notanemail","password":"secret123"}`,
			wantStatus: 400,
		},
		{
			name:       "missing password",
			body:       `{"name":"John","email":"john3@example.com"}`,
			wantStatus: 400,
		},
		{
			name:       "password too short",
			body:       `{"name":"John","email":"john4@example.com","password":"abc"}`,
			wantStatus: 400,
		},
		{
			name:       "malformed JSON",
			body:       `{bad json}`,
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupAuthTestEnv(t)
			defer cleanup()

			c := newAuthController(http.MethodPost, "/api/v1/auth/register", []byte(tt.body))
			c.Register()

			status := responseStatus(c)
			if status != tt.wantStatus {
				t.Errorf("Register() status=%d, want %d", status, tt.wantStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid login",
			body:       `{"email":"existing@example.com","password":"pass123"}`,
			wantStatus: 200,
		},
		{
			name:       "wrong password",
			body:       `{"email":"existing@example.com","password":"wrongpass"}`,
			wantStatus: 401,
		},
		{
			name:       "non-existing email",
			body:       `{"email":"nobody@example.com","password":"pass123"}`,
			wantStatus: 401,
		},
		{
			name:       "missing email",
			body:       `{"password":"pass123"}`,
			wantStatus: 400,
		},
		{
			name:       "missing password",
			body:       `{"email":"existing@example.com"}`,
			wantStatus: 400,
		},
		{
			name:       "malformed JSON",
			body:       `{bad json}`,
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupAuthTestEnv(t)
			defer cleanup()

			c := newAuthController(http.MethodPost, "/api/v1/auth/login", []byte(tt.body))
			c.Login()

			status := responseStatus(c)
			if status != tt.wantStatus {
				t.Errorf("Login() status=%d, want %d", status, tt.wantStatus)
			}
		})
	}
}
