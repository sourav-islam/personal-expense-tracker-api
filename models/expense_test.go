package models

import (
	"os"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// setupExpenseTestCSV sets a temp path for expenses CSV in config.
func setupExpenseTestCSV(t *testing.T) func() {
	t.Helper()
	// Also need a users CSV path so user lookups don't fail
	userTmp := t.TempDir() + "/users_test.csv"
	expTmp := t.TempDir() + "/expenses_test.csv"
	beego.AppConfig.Set("users_csv_path", userTmp)
	beego.AppConfig.Set("expenses_csv_path", expTmp)
	return func() {
		os.Remove(userTmp)
		os.Remove(expTmp)
	}
}

// sampleExpense returns a valid Expense for user 1.
func sampleExpense(title, category, date string, amount float64) *Expense {
	return &Expense{
		UserID:      1,
		Title:       title,
		Amount:      amount,
		Category:    category,
		Note:        "test note",
		ExpenseDate: date,
	}
}

// ---------------------------------------------------------------------------
// IsValidCategory
// ---------------------------------------------------------------------------

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     bool
	}{
		{name: "Food valid", category: "Food", want: true},
		{name: "Transport valid", category: "Transport", want: true},
		{name: "Housing valid", category: "Housing", want: true},
		{name: "Entertainment valid", category: "Entertainment", want: true},
		{name: "Shopping valid", category: "Shopping", want: true},
		{name: "Healthcare valid", category: "Healthcare", want: true},
		{name: "Education valid", category: "Education", want: true},
		{name: "Utilities valid", category: "Utilities", want: true},
		{name: "Other valid", category: "Other", want: true},
		{name: "invalid category", category: "Snacks", want: false},
		{name: "empty string", category: "", want: false},
		{name: "lowercase food", category: "food", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidCategory(tt.category)
			if got != tt.want {
				t.Errorf("IsValidCategory(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsValidDateFormat
// ---------------------------------------------------------------------------

func TestIsValidDateFormat(t *testing.T) {
	tests := []struct {
		name string
		date string
		want bool
	}{
		{name: "valid date", date: "2025-06-01", want: true},
		{name: "valid date end of year", date: "2025-12-31", want: true},
		{name: "wrong separator", date: "2025/06/01", want: false},
		{name: "day first format", date: "01-06-2025", want: false},
		{name: "missing day", date: "2025-06", want: false},
		{name: "empty string", date: "", want: false},
		{name: "invalid month 13", date: "2025-13-01", want: false},
		{name: "invalid day 32", date: "2025-06-32", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDateFormat(tt.date)
			if got != tt.want {
				t.Errorf("IsValidDateFormat(%q) = %v, want %v", tt.date, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CreateExpense + GetExpensesByUserID
// ---------------------------------------------------------------------------

func TestCreateExpense_Success(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	e := sampleExpense("Lunch", "Food", "2025-06-01", 350.50)
	if err := CreateExpense(e); err != nil {
		t.Fatalf("CreateExpense() unexpected error: %v", err)
	}
	if e.ID != 1 {
		t.Errorf("expected ID=1, got %d", e.ID)
	}
	if e.CreatedAt == "" {
		t.Error("expected CreatedAt to be populated")
	}
}

func TestGetExpensesByUserID_OnlyOwnExpenses(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	// User 1 expenses
	_ = CreateExpense(sampleExpense("Lunch", "Food", "2025-06-01", 100))
	_ = CreateExpense(sampleExpense("Bus", "Transport", "2025-06-02", 50))

	// User 2 expense — must not appear in user 1 results
	e3 := sampleExpense("Dinner", "Food", "2025-06-03", 200)
	e3.UserID = 2
	_ = CreateExpense(e3)

	expenses, err := GetExpensesByUserID(1)
	if err != nil {
		t.Fatalf("GetExpensesByUserID() error: %v", err)
	}
	if len(expenses) != 2 {
		t.Errorf("expected 2 expenses for user 1, got %d", len(expenses))
	}
	for _, ex := range expenses {
		if ex.UserID != 1 {
			t.Errorf("got expense with UserID=%d, expected 1", ex.UserID)
		}
	}
}

func TestGetExpensesByUserID_Empty(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	expenses, err := GetExpensesByUserID(99)
	if err != nil {
		t.Fatalf("GetExpensesByUserID() error: %v", err)
	}
	if len(expenses) != 0 {
		t.Errorf("expected 0 expenses, got %d", len(expenses))
	}
}

// ---------------------------------------------------------------------------
// GetExpenseByID
// ---------------------------------------------------------------------------

func TestGetExpenseByID(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	e := sampleExpense("Lunch", "Food", "2025-06-01", 350.50)
	_ = CreateExpense(e)

	tests := []struct {
		name      string
		id        int
		userID    int
		wantFound bool
	}{
		{name: "correct ID and owner", id: 1, userID: 1, wantFound: true},
		{name: "correct ID wrong owner", id: 1, userID: 2, wantFound: false},
		{name: "non-existing ID", id: 999, userID: 1, wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense, err := GetExpenseByID(tt.id, tt.userID)
			if err != nil {
				t.Fatalf("GetExpenseByID() unexpected error: %v", err)
			}
			found := expense != nil
			if found != tt.wantFound {
				t.Errorf("GetExpenseByID(%d,%d) found=%v, want %v",
					tt.id, tt.userID, found, tt.wantFound)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateExpense
// ---------------------------------------------------------------------------

func TestUpdateExpense_Success(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	e := sampleExpense("Lunch", "Food", "2025-06-01", 350.50)
	_ = CreateExpense(e)

	e.Title = "Dinner"
	e.Amount = 600.00
	if err := UpdateExpense(e); err != nil {
		t.Fatalf("UpdateExpense() unexpected error: %v", err)
	}

	updated, _ := GetExpenseByID(e.ID, 1)
	if updated == nil {
		t.Fatal("expense not found after update")
	}
	if updated.Title != "Dinner" {
		t.Errorf("expected title=Dinner, got %s", updated.Title)
	}
	if updated.Amount != 600.00 {
		t.Errorf("expected amount=600.00, got %.2f", updated.Amount)
	}
}

func TestUpdateExpense_NotFound(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	e := &Expense{ID: 999, UserID: 1, Title: "Ghost",
		Amount: 1, Category: "Food", ExpenseDate: "2025-06-01"}
	err := UpdateExpense(e)
	if err == nil {
		t.Error("expected error for non-existing expense, got nil")
	}
}

// ---------------------------------------------------------------------------
// DeleteExpense
// ---------------------------------------------------------------------------

func TestDeleteExpense_Success(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	e := sampleExpense("Lunch", "Food", "2025-06-01", 350.50)
	_ = CreateExpense(e)

	if err := DeleteExpense(e.ID, 1); err != nil {
		t.Fatalf("DeleteExpense() unexpected error: %v", err)
	}

	remaining, _ := GetExpensesByUserID(1)
	if len(remaining) != 0 {
		t.Errorf("expected 0 expenses after delete, got %d", len(remaining))
	}
}

func TestDeleteExpense_WrongOwner(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	e := sampleExpense("Lunch", "Food", "2025-06-01", 350.50)
	_ = CreateExpense(e)

	err := DeleteExpense(e.ID, 2) // wrong userID
	if err == nil {
		t.Error("expected error when deleting another user's expense")
	}
}

func TestDeleteExpense_NotFound(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	err := DeleteExpense(999, 1)
	if err == nil {
		t.Error("expected error for non-existing expense ID")
	}
}

// ---------------------------------------------------------------------------
// GetNextExpenseID
// ---------------------------------------------------------------------------

func TestGetNextExpenseID(t *testing.T) {
	cleanup := setupExpenseTestCSV(t)
	defer cleanup()

	id, err := GetNextExpenseID()
	if err != nil {
		t.Fatalf("GetNextExpenseID() error: %v", err)
	}
	if id != 1 {
		t.Errorf("expected 1 on empty file, got %d", id)
	}

	_ = CreateExpense(sampleExpense("Lunch", "Food", "2025-06-01", 100))
	id, err = GetNextExpenseID()
	if err != nil {
		t.Fatalf("GetNextExpenseID() error: %v", err)
	}
	if id != 2 {
		t.Errorf("expected 2 after one record, got %d", id)
	}
}
