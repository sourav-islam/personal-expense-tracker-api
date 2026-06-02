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

// setupExpenseControllerEnv creates temp CSVs and seeds one user + one expense.
func setupExpenseControllerEnv(t *testing.T) func() {
	t.Helper()
	userTmp := t.TempDir() + "/users_test.csv"
	expTmp := t.TempDir() + "/expenses_test.csv"
	beego.AppConfig.Set("users_csv_path", userTmp)
	beego.AppConfig.Set("expenses_csv_path", expTmp)

	_ = models.CreateUser(&models.User{
		Name: "Alice", Email: "alice@example.com", Password: "pass123",
	})
	_ = models.CreateExpense(&models.Expense{
		UserID: 1, Title: "Lunch", Amount: 350.50,
		Category: "Food", ExpenseDate: "2025-06-01",
	})
	return func() {
		os.Remove(userTmp)
		os.Remove(expTmp)
	}
}

// newExpenseController builds an ExpenseController wired to a fake request.
func newExpenseController(
	method, path, userIDHeader string,
	body []byte,
	pathParams map[string]string,
) *ExpenseController {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if userIDHeader != "" {
		req.Header.Set("X-User-ID", userIDHeader)
	}
	rw := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(rw, req)
	ctx.Input.RequestBody = body
	for k, v := range pathParams {
		ctx.Input.SetParam(k, v)
	}

	c := &ExpenseController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	return c
}

func expenseResponseStatus(c *ExpenseController) int {
	return c.Ctx.ResponseWriter.Status
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestExpenseCreate(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       string
		wantStatus int
	}{
		{
			name:       "valid expense",
			userID:     "1",
			body:       `{"title":"Lunch","amount":350.50,"category":"Food","expense_date":"2025-06-10"}`,
			wantStatus: 201,
		},
		{
			name:       "missing title",
			userID:     "1",
			body:       `{"amount":350.50,"category":"Food","expense_date":"2025-06-10"}`,
			wantStatus: 400,
		},
		{
			name:       "zero amount",
			userID:     "1",
			body:       `{"title":"Lunch","amount":0,"category":"Food","expense_date":"2025-06-10"}`,
			wantStatus: 400,
		},
		{
			name:       "negative amount",
			userID:     "1",
			body:       `{"title":"Lunch","amount":-100,"category":"Food","expense_date":"2025-06-10"}`,
			wantStatus: 400,
		},
		{
			name:       "invalid category",
			userID:     "1",
			body:       `{"title":"Lunch","amount":100,"category":"Snacks","expense_date":"2025-06-10"}`,
			wantStatus: 400,
		},
		{
			name:       "missing expense_date",
			userID:     "1",
			body:       `{"title":"Lunch","amount":100,"category":"Food"}`,
			wantStatus: 400,
		},
		{
			name:       "invalid date format",
			userID:     "1",
			body:       `{"title":"Lunch","amount":100,"category":"Food","expense_date":"10-06-2025"}`,
			wantStatus: 400,
		},
		{
			name:       "no auth header",
			userID:     "",
			body:       `{"title":"Lunch","amount":100,"category":"Food","expense_date":"2025-06-10"}`,
			wantStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseControllerEnv(t)
			defer cleanup()

			c := newExpenseController(http.MethodPost, "/api/v1/expenses",
				tt.userID, []byte(tt.body), nil)
			c.Create()

			if got := expenseResponseStatus(c); got != tt.wantStatus {
				t.Errorf("Create() status=%d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetOne
// ---------------------------------------------------------------------------

func TestExpenseGetOne(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		expenseID  string
		wantStatus int
	}{
		{name: "found", userID: "1", expenseID: "1", wantStatus: 200},
		{name: "not found", userID: "1", expenseID: "999", wantStatus: 404},
		{name: "wrong owner", userID: "2", expenseID: "1", wantStatus: 401},
		{name: "invalid id", userID: "1", expenseID: "abc", wantStatus: 400},
		{name: "no auth", userID: "", expenseID: "1", wantStatus: 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseControllerEnv(t)
			defer cleanup()

			c := newExpenseController(http.MethodGet, "/api/v1/expenses/"+tt.expenseID,
				tt.userID, nil, map[string]string{":id": tt.expenseID})
			c.GetOne()

			if got := expenseResponseStatus(c); got != tt.wantStatus {
				t.Errorf("GetOne() status=%d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestExpenseUpdate(t *testing.T) {
	validBody := `{"title":"Dinner","amount":500.00,"category":"Food","expense_date":"2025-06-10"}`

	tests := []struct {
		name       string
		userID     string
		expenseID  string
		body       string
		wantStatus int
	}{
		{name: "valid update", userID: "1", expenseID: "1", body: validBody, wantStatus: 200},
		{name: "not found", userID: "1", expenseID: "999", body: validBody, wantStatus: 404},
		{name: "wrong owner", userID: "2", expenseID: "1", body: validBody, wantStatus: 401},
		{
			name: "invalid body", userID: "1", expenseID: "1",
			body:       `{"title":"","amount":0,"category":"Food","expense_date":"2025-06-10"}`,
			wantStatus: 400,
		},
		{name: "no auth", userID: "", expenseID: "1", body: validBody, wantStatus: 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseControllerEnv(t)
			defer cleanup()

			c := newExpenseController(http.MethodPut, "/api/v1/expenses/"+tt.expenseID,
				tt.userID, []byte(tt.body), map[string]string{":id": tt.expenseID})
			c.Update()

			if got := expenseResponseStatus(c); got != tt.wantStatus {
				t.Errorf("Update() status=%d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestExpenseDelete(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		expenseID  string
		wantStatus int
	}{
		{name: "valid delete", userID: "1", expenseID: "1", wantStatus: 200},
		{name: "not found", userID: "1", expenseID: "999", wantStatus: 404},
		{name: "wrong owner", userID: "2", expenseID: "1", wantStatus: 401},
		{name: "invalid id", userID: "1", expenseID: "abc", wantStatus: 400},
		{name: "no auth", userID: "", expenseID: "1", wantStatus: 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseControllerEnv(t)
			defer cleanup()

			c := newExpenseController(http.MethodDelete, "/api/v1/expenses/"+tt.expenseID,
				tt.userID, nil, map[string]string{":id": tt.expenseID})
			c.Delete()

			if got := expenseResponseStatus(c); got != tt.wantStatus {
				t.Errorf("Delete() status=%d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Summary
// ---------------------------------------------------------------------------

func TestExpenseSummary(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		query      string
		wantStatus int
	}{
		{
			name: "valid summary", userID: "1",
			query:      "?date_from=2025-06-01&date_to=2025-06-30",
			wantStatus: 200,
		},
		{
			name: "missing date_from", userID: "1",
			query:      "?date_to=2025-06-30",
			wantStatus: 400,
		},
		{
			name: "missing date_to", userID: "1",
			query:      "?date_from=2025-06-01",
			wantStatus: 400,
		},
		{
			name: "date_from after date_to", userID: "1",
			query:      "?date_from=2025-07-01&date_to=2025-06-01",
			wantStatus: 400,
		},
		{
			name: "invalid date_from format", userID: "1",
			query:      "?date_from=01-06-2025&date_to=2025-06-30",
			wantStatus: 400,
		},
		{
			name:       "no auth",
			userID:     "",
			query:      "?date_from=2025-06-01&date_to=2025-06-30",
			wantStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseControllerEnv(t)
			defer cleanup()

			c := newExpenseController(http.MethodGet,
				"/api/v1/expenses/summary"+tt.query,
				tt.userID, nil, nil)
			c.Summary()

			if got := expenseResponseStatus(c); got != tt.wantStatus {
				t.Errorf("Summary() status=%d, want %d", got, tt.wantStatus)
			}
		})
	}
}
