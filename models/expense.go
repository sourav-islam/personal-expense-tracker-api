// Package models handles all data structures and CSV file operations.
package models

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

// Expense represents a single expense record belonging to a user.
type Expense struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
	CreatedAt   string  `json:"created_at"`
}

// AllowedCategories contains the valid category values for an expense.
var AllowedCategories = []string{
	"Food", "Transport", "Housing", "Entertainment",
	"Shopping", "Healthcare", "Education", "Utilities", "Other",
}

// expenseCSVHeader is the fixed header row for expenses.csv.
var expenseCSVHeader = []string{
	"id", "user_id", "title", "amount",
	"category", "note", "expense_date", "created_at",
}

// getExpensesCSVPath returns the file path for expenses.csv from config.
func getExpensesCSVPath() string {
	return beego.AppConfig.DefaultString("expenses_csv_path", "data/expenses.csv")
}

// ensureExpensesCSV creates the expenses CSV file with header if it does not exist.
func ensureExpensesCSV(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
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
		if err := w.Write(expenseCSVHeader); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	}
	return nil
}

// readAllExpenseRecords reads every row from the expenses CSV as raw string slices.
// Returns rows including the header row at index 0.
func readAllExpenseRecords(path string) ([][]string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	return reader.ReadAll()
}

// writeAllExpenseRecords overwrites the entire expenses CSV with the given rows.
// The header row must be included as the first element.
func writeAllExpenseRecords(path string, records [][]string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.WriteAll(records); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// recordToExpense converts a CSV string slice into an Expense struct.
// Returns an error if any required field cannot be parsed.
func recordToExpense(record []string) (Expense, error) {
	id, err := strconv.Atoi(record[0])
	if err != nil {
		return Expense{}, err
	}
	userID, err := strconv.Atoi(record[1])
	if err != nil {
		return Expense{}, err
	}
	amount, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return Expense{}, err
	}
	return Expense{
		ID:          id,
		UserID:      userID,
		Title:       record[2],
		Amount:      amount,
		Category:    record[4],
		Note:        record[5],
		ExpenseDate: record[6],
		CreatedAt:   record[7],
	}, nil
}

// expenseToRecord converts an Expense struct into a CSV string slice.
func expenseToRecord(e *Expense) []string {
	return []string{
		strconv.Itoa(e.ID),
		strconv.Itoa(e.UserID),
		e.Title,
		strconv.FormatFloat(e.Amount, 'f', 2, 64),
		e.Category,
		e.Note,
		e.ExpenseDate,
		e.CreatedAt,
	}
}

// GetNextExpenseID returns the next available expense ID based on
// the highest existing ID in the CSV file.
func GetNextExpenseID() (int, error) {
	path := getExpensesCSVPath()
	if err := ensureExpensesCSV(path); err != nil {
		return 0, err
	}

	records, err := readAllExpenseRecords(path)
	if err != nil {
		return 0, err
	}

	maxID := 0
	// Skip header at index 0
	for _, record := range records[1:] {
		if len(record) < 1 {
			continue
		}
		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1, nil
}

// GetExpensesByUserID returns all expenses belonging to the given user ID.
func GetExpensesByUserID(userID int) ([]Expense, error) {
	path := getExpensesCSVPath()
	if err := ensureExpensesCSV(path); err != nil {
		logs.Error("Failed to ensure expenses CSV:", err)
		return nil, err
	}

	records, err := readAllExpenseRecords(path)
	if err != nil {
		logs.Error("Failed to read expenses CSV:", err)
		return nil, err
	}

	var expenses []Expense
	for _, record := range records[1:] {
		if len(record) < 8 {
			logs.Warn("Skipping malformed expense row:", record)
			continue
		}
		expense, err := recordToExpense(record)
		if err != nil {
			logs.Warn("Skipping expense row with parse error:", err)
			continue
		}
		if expense.UserID == userID {
			expenses = append(expenses, expense)
		}
	}

	return expenses, nil
}

// GetExpenseByID returns a single expense matching both the expense ID
// and the user ID. Returns nil if not found or if ownership does not match.
func GetExpenseByID(id int, userID int) (*Expense, error) {
	expenses, err := GetExpensesByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, e := range expenses {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, nil
}

// CreateExpense appends a new expense record to the CSV file.
// It assigns the next available ID and sets CreatedAt to the current UTC time.
func CreateExpense(expense *Expense) error {
	path := getExpensesCSVPath()
	if err := ensureExpensesCSV(path); err != nil {
		logs.Error("Failed to ensure expenses CSV before creating expense:", err)
		return err
	}

	nextID, err := GetNextExpenseID()
	if err != nil {
		return err
	}
	expense.ID = nextID
	expense.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logs.Error("Failed to open expenses CSV for append:", err)
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(expenseToRecord(expense)); err != nil {
		logs.Error("Failed to write expense record:", err)
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		logs.Error("CSV writer flush error:", err)
		return err
	}

	logs.Info("Created expense ID:", expense.ID, "UserID:", expense.UserID)
	return nil
}

// UpdateExpense replaces an existing expense record in the CSV file.
// It rewrites the entire file with the updated row in place.
func UpdateExpense(updated *Expense) error {
	path := getExpensesCSVPath()

	// Ensure file exists before attempting to read
	if err := ensureExpensesCSV(path); err != nil {
		logs.Error("Failed to ensure expenses CSV before update:", err)
		return err
	}

	records, err := readAllExpenseRecords(path)
	if err != nil {
		logs.Error("Failed to read expenses CSV for update:", err)
		return err
	}

	found := false
	for i, record := range records[1:] {
		if len(record) < 1 {
			continue
		}
		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		if id == updated.ID {
			updated.CreatedAt = record[7]
			records[i+1] = expenseToRecord(updated)
			found = true
			break
		}
	}

	if !found {
		logs.Warn("UpdateExpense: expense ID not found:", updated.ID)
		return ErrUserNotFound
	}

	if err := writeAllExpenseRecords(path, records); err != nil {
		logs.Error("Failed to write updated expenses CSV:", err)
		return err
	}

	logs.Info("Updated expense ID:", updated.ID)
	return nil
}

// DeleteExpense removes an expense by ID from the CSV file,
// only if it belongs to the given user ID.
func DeleteExpense(id int, userID int) error {
	path := getExpensesCSVPath()

	// Ensure file exists before attempting to read
	if err := ensureExpensesCSV(path); err != nil {
		logs.Error("Failed to ensure expenses CSV before delete:", err)
		return err
	}

	records, err := readAllExpenseRecords(path)
	if err != nil {
		logs.Error("Failed to read expenses CSV for delete:", err)
		return err
	}

	newRecords := [][]string{records[0]}
	found := false

	for _, record := range records[1:] {
		if len(record) < 2 {
			continue
		}
		expID, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		expUserID, err := strconv.Atoi(record[1])
		if err != nil {
			continue
		}
		if expID == id && expUserID == userID {
			found = true
			continue
		}
		newRecords = append(newRecords, record)
	}

	if !found {
		logs.Warn("DeleteExpense: expense ID not found or ownership mismatch:", id)
		return ErrUserNotFound
	}

	if err := writeAllExpenseRecords(path, newRecords); err != nil {
		logs.Error("Failed to write expenses CSV after delete:", err)
		return err
	}

	logs.Info("Deleted expense ID:", id, "UserID:", userID)
	return nil
}

// IsValidCategory checks whether the given category string is in AllowedCategories.
func IsValidCategory(category string) bool {
	for _, c := range AllowedCategories {
		if c == category {
			return true
		}
	}
	return false
}

// IsValidDateFormat checks whether the given string is in YYYY-MM-DD format.
func IsValidDateFormat(date string) bool {
	if len(date) != 10 {
		return false
	}
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return false
	}
	if len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return false
	}
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
