package models

import (
	"os"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// setupUserTestCSV creates a temp CSV path in config and removes
// any leftover file before each test.
func setupUserTestCSV(t *testing.T) func() {
	t.Helper()
	tmpPath := t.TempDir() + "/users_test.csv"
	beego.AppConfig.Set("users_csv_path", tmpPath)
	return func() {
		os.Remove(tmpPath)
	}
}

// ---------------------------------------------------------------------------
// ValidateEmail
// ---------------------------------------------------------------------------

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "valid email", email: "john@example.com", want: true},
		{name: "valid with subdomain", email: "john@mail.example.com", want: true},
		{name: "missing @", email: "johnexample.com", want: false},
		{name: "missing domain dot", email: "john@example", want: false},
		{name: "empty string", email: "", want: false},
		{name: "only @", email: "@", want: false},
		{name: "no local part", email: "@example.com", want: false},
		{name: "no domain part", email: "john@", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got != tt.want {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetAllUsers — empty file
// ---------------------------------------------------------------------------

func TestGetAllUsers_EmptyFile(t *testing.T) {
	cleanup := setupUserTestCSV(t)
	defer cleanup()

	users, err := GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers() unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

// ---------------------------------------------------------------------------
// CreateUser + GetAllUsers
// ---------------------------------------------------------------------------

func TestCreateUser_Success(t *testing.T) {
	cleanup := setupUserTestCSV(t)
	defer cleanup()

	user := &User{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "pass123",
	}

	if err := CreateUser(user); err != nil {
		t.Fatalf("CreateUser() unexpected error: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected ID=1, got %d", user.ID)
	}
	if user.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}

	users, err := GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers() unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", users[0].Email)
	}
}

func TestCreateUser_MultipleUsers_IDIncrement(t *testing.T) {
	cleanup := setupUserTestCSV(t)
	defer cleanup()

	for i, name := range []string{"Alice", "Bob", "Carol"} {
		u := &User{Name: name, Email: name + "@example.com", Password: "pass123"}
		if err := CreateUser(u); err != nil {
			t.Fatalf("CreateUser() error: %v", err)
		}
		if u.ID != i+1 {
			t.Errorf("expected ID=%d, got %d", i+1, u.ID)
		}
	}
}

// ---------------------------------------------------------------------------
// GetUserByEmail
// ---------------------------------------------------------------------------

func TestGetUserByEmail(t *testing.T) {
	cleanup := setupUserTestCSV(t)
	defer cleanup()

	_ = CreateUser(&User{Name: "Alice", Email: "alice@example.com", Password: "pass123"})

	tests := []struct {
		name      string
		email     string
		wantFound bool
	}{
		{name: "existing email", email: "alice@example.com", wantFound: true},
		{name: "non-existing email", email: "nobody@example.com", wantFound: false},
		{name: "empty email", email: "", wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := GetUserByEmail(tt.email)
			if err != nil {
				t.Fatalf("GetUserByEmail() unexpected error: %v", err)
			}
			found := user != nil
			if found != tt.wantFound {
				t.Errorf("GetUserByEmail(%q) found=%v, want %v", tt.email, found, tt.wantFound)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserByID
// ---------------------------------------------------------------------------

func TestGetUserByID(t *testing.T) {
	cleanup := setupUserTestCSV(t)
	defer cleanup()

	_ = CreateUser(&User{Name: "Alice", Email: "alice@example.com", Password: "pass123"})

	tests := []struct {
		name      string
		id        int
		wantFound bool
	}{
		{name: "existing ID", id: 1, wantFound: true},
		{name: "non-existing ID", id: 999, wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := GetUserByID(tt.id)
			if err != nil {
				t.Fatalf("GetUserByID() unexpected error: %v", err)
			}
			found := user != nil
			if found != tt.wantFound {
				t.Errorf("GetUserByID(%d) found=%v, want %v", tt.id, found, tt.wantFound)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetNextUserID
// ---------------------------------------------------------------------------

func TestGetNextUserID(t *testing.T) {
	cleanup := setupUserTestCSV(t)
	defer cleanup()

	// Empty file → next ID is 1
	id, err := GetNextUserID()
	if err != nil {
		t.Fatalf("GetNextUserID() error: %v", err)
	}
	if id != 1 {
		t.Errorf("expected next ID=1, got %d", id)
	}

	// After creating one user → next ID is 2
	_ = CreateUser(&User{Name: "Alice", Email: "alice@example.com", Password: "pass123"})
	id, err = GetNextUserID()
	if err != nil {
		t.Fatalf("GetNextUserID() error: %v", err)
	}
	if id != 2 {
		t.Errorf("expected next ID=2, got %d", id)
	}
}
