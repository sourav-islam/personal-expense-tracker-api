// Package models handles all data structures and CSV file operations.
package models

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

// User represents a registered user in the system.
type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

// csvHeader is the fixed header row for users.csv.
var csvHeader = []string{"id", "name", "email", "password", "created_at"}

// getUsersCSVPath returns the file path for users.csv from config.
func getUsersCSVPath() string {
	return beego.AppConfig.DefaultString("users_csv_path", "data/users.csv")
}

// ensureUsersCSV creates the users CSV file with header if it does not exist.
func ensureUsersCSV(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Ensure parent directory exists
		dir := path[:strings.LastIndex(path, "/")]
		if dir != "" {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		w := csv.NewWriter(f)
		if err := w.Write(csvHeader); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	}
	return nil
}

// GetAllUsers reads and returns all users from the CSV file.
func GetAllUsers() ([]User, error) {
	path := getUsersCSVPath()
	if err := ensureUsersCSV(path); err != nil {
		logs.Error("Failed to ensure users CSV:", err)
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		logs.Error("Failed to open users CSV:", err)
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		logs.Error("Failed to read users CSV:", err)
		return nil, err
	}

	var users []User
	// Skip header row (index 0)
	for _, record := range records[1:] {
		if len(record) < 5 {
			logs.Warn("Skipping malformed user row:", record)
			continue
		}
		id, err := strconv.Atoi(record[0])
		if err != nil {
			logs.Warn("Skipping user row with invalid ID:", record[0])
			continue
		}
		users = append(users, User{
			ID:        id,
			Name:      record[1],
			Email:     record[2],
			Password:  record[3],
			CreatedAt: record[4],
		})
	}

	return users, nil
}

// GetUserByEmail finds and returns a user matching the given email.
// Returns nil and no error if the user is not found.
func GetUserByEmail(email string) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if strings.EqualFold(u.Email, email) {
			return &u, nil
		}
	}
	return nil, nil
}

// GetUserByID finds and returns a user matching the given ID.
// Returns nil and no error if the user is not found.
func GetUserByID(id int) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

// GetNextUserID returns the next available user ID based on
// the highest existing ID in the CSV file.
func GetNextUserID() (int, error) {
	users, err := GetAllUsers()
	if err != nil {
		return 0, err
	}
	maxID := 0
	for _, u := range users {
		if u.ID > maxID {
			maxID = u.ID
		}
	}
	return maxID + 1, nil
}

// CreateUser appends a new user record to the CSV file.
// It creates the file with headers if it does not already exist.
func CreateUser(user *User) error {
	path := getUsersCSVPath()
	if err := ensureUsersCSV(path); err != nil {
		logs.Error("Failed to ensure users CSV before creating user:", err)
		return err
	}

	nextID, err := GetNextUserID()
	if err != nil {
		return err
	}
	user.ID = nextID
	user.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logs.Error("Failed to open users CSV for append:", err)
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	record := []string{
		strconv.Itoa(user.ID),
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
	}
	if err := w.Write(record); err != nil {
		logs.Error("Failed to write user record:", err)
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		logs.Error("CSV writer flush error:", err)
		return err
	}

	logs.Info("Created user ID:", user.ID, "Email:", user.Email)
	return nil
}

// ValidateEmail checks whether the given string is a valid email format.
func ValidateEmail(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// ErrUserNotFound is returned when a requested user does not exist.
var ErrUserNotFound = errors.New("user not found")
