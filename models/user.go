package models

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// User represents a user in the system.
type User struct {
	ID        int
	Name      string
	Email     string
	Password  string
	CreatedAt string
}

// GetAllUsers reads all users from the CSV file.
func GetAllUsers() ([]User, error) {
	filePath, err := web.AppConfig.String("users_csv_path")
	if err != nil {
		logs.Error("Failed to get users_csv_path from config: %v", err)
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		logs.Error("Failed to open users CSV file: %v", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	_, err = reader.Read()
	if err != nil {
		if err == io.EOF {
			return []User{}, nil
		}
		logs.Error("Failed to read header from users CSV: %v", err)
		return nil, err
	}

	var users []User
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logs.Error("Failed to read record from users CSV: %v", err)
			return nil, err
		}

		id, _ := strconv.Atoi(record[0])
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

// GetUserByEmail finds a user by their email address.
func GetUserByEmail(email string) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, nil
}

// CreateUser appends a new user to the CSV file.
func CreateUser(user *User) error {
	filePath, err := web.AppConfig.String("users_csv_path")
	if err != nil {
		logs.Error("Failed to get users_csv_path from config: %v", err)
		return err
	}

	fileExists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logs.Error("Failed to open users CSV file for writing: %v", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		header := []string{"id", "name", "email", "password", "created_at"}
		if err := writer.Write(header); err != nil {
			logs.Error("Failed to write header to users CSV: %v", err)
			return err
		}
	}

	record := []string{
		strconv.Itoa(user.ID),
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
	}

	if err := writer.Write(record); err != nil {
		logs.Error("Failed to write user record to CSV: %v", err)
		return err
	}

	return nil
}

// GetNextUserID returns the next available user ID.
func GetNextUserID() (int, error) {
	users, err := GetAllUsers()
	if err != nil {
		return 0, err
	}

	maxID := 0
	for _, user := range users {
		if user.ID > maxID {
			maxID = user.ID
		}
	}

	return maxID + 1, nil
}
