package controllers

import (
	"encoding/json"
	"expense-tracker-api/models"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
)

// ExpenseController handles all expense CRUD operations.
type ExpenseController struct {
	BaseController
}

// createExpenseInput defines the expected JSON body for creating an expense.
type createExpenseInput struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
}

// updateExpenseInput defines the expected JSON body for updating an expense.
type updateExpenseInput struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
}

// validateExpenseInput checks all required fields for create/update operations.
// Returns an error message string, or empty string if valid.
func validateExpenseInput(title, category, expenseDate string, amount float64) string {
	if title == "" {
		return "Title is required"
	}
	if amount <= 0 {
		return "Amount must be a positive number"
	}
	if expenseDate == "" {
		return "Expense date is required"
	}
	if !models.IsValidDateFormat(expenseDate) {
		return "Invalid expense_date format, expected YYYY-MM-DD"
	}
	if category == "" {
		return "Category is required"
	}
	if !models.IsValidCategory(category) {
		return "Invalid category"
	}
	return ""
}

// Create godoc
// @Title Create Expense
// @Summary Create a new expense for the authenticated user
// @Param X-User-ID header int true "User ID"
// @Param body body controllers.createExpenseInput true "Expense payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses [post]
func (c *ExpenseController) Create() {
	logs.Info("Create expense endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	var input createExpenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Create expense: failed to parse request body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if errMsg := validateExpenseInput(input.Title, input.Category, input.ExpenseDate, input.Amount); errMsg != "" {
		logs.Warn("Create expense: invalid input:", errMsg)
		c.SendError(400, errMsg)
		return
	}

	expense := &models.Expense{
		UserID:      userID,
		Title:       input.Title,
		Amount:      input.Amount,
		Category:    input.Category,
		Note:        input.Note,
		ExpenseDate: input.ExpenseDate,
	}

	if err := models.CreateExpense(expense); err != nil {
		logs.Error("Create expense: failed to save expense:", err)
		c.SendError(500, "Failed to create expense")
		return
	}

	logs.Info("Create expense: success, ID:", expense.ID)
	c.SendSuccess(201, "Expense created successfully", expense)
}

// List godoc
// @Title List Expenses
// @Summary List all expenses for the authenticated user
// @Param X-User-ID header int true "User ID"
// @Param limit query int false "Max number of results"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses [get]
func (c *ExpenseController) List() {
	logs.Info("List expenses endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("List expenses: failed to retrieve expenses:", err)
		c.SendError(500, "Failed to retrieve expenses")
		return
	}

	// Apply limit if provided
	limitStr := c.GetString("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			logs.Warn("List expenses: invalid limit parameter:", limitStr)
			c.SendError(400, "Invalid limit parameter")
			return
		}
		if limit < len(expenses) {
			expenses = expenses[:limit]
		}
	}

	// Return empty array instead of null
	if expenses == nil {
		expenses = []models.Expense{}
	}

	logs.Info("List expenses: returning", len(expenses), "expenses for user:", userID)
	c.SendSuccess(200, "Expenses retrieved", expenses)
}

// GetOne godoc
// @Title Get Expense
// @Summary Get a single expense by ID
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses/:id [get]
func (c *ExpenseController) GetOne() {
	logs.Info("Get one expense endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	expenseID, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || expenseID <= 0 {
		logs.Warn("GetOne: invalid expense ID param:", c.Ctx.Input.Param(":id"))
		c.SendError(400, "Invalid expense ID")
		return
	}

	expense, err := models.GetExpenseByID(expenseID, userID)
	if err != nil {
		logs.Error("GetOne: failed to retrieve expense:", err)
		c.SendError(500, "Failed to retrieve expense")
		return
	}
	if expense == nil {
		logs.Warn("GetOne: expense not found, ID:", expenseID, "UserID:", userID)
		c.SendError(404, "Expense not found")
		return
	}

	logs.Info("GetOne: found expense ID:", expenseID)
	c.SendSuccess(200, "Expense retrieved", expense)
}

// Update godoc
// @Title Update Expense
// @Summary Update an existing expense by ID
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Param body body controllers.updateExpenseInput true "Updated expense payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses/:id [put]
func (c *ExpenseController) Update() {
	logs.Info("Update expense endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	expenseID, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || expenseID <= 0 {
		logs.Warn("Update: invalid expense ID param:", c.Ctx.Input.Param(":id"))
		c.SendError(400, "Invalid expense ID")
		return
	}

	// Confirm ownership before updating
	existing, err := models.GetExpenseByID(expenseID, userID)
	if err != nil {
		logs.Error("Update: error fetching expense:", err)
		c.SendError(500, "Failed to update expense")
		return
	}
	if existing == nil {
		logs.Warn("Update: expense not found or not owned, ID:", expenseID, "UserID:", userID)
		c.SendError(404, "Expense not found")
		return
	}

	var input updateExpenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Update: failed to parse request body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if errMsg := validateExpenseInput(input.Title, input.Category, input.ExpenseDate, input.Amount); errMsg != "" {
		c.SendError(400, errMsg)
		return
	}

	updated := &models.Expense{
		ID:          expenseID,
		UserID:      userID,
		Title:       input.Title,
		Amount:      input.Amount,
		Category:    input.Category,
		Note:        input.Note,
		ExpenseDate: input.ExpenseDate,
	}

	if err := models.UpdateExpense(updated); err != nil {
		logs.Error("Update: failed to update expense:", err)
		c.SendError(500, "Failed to update expense")
		return
	}

	logs.Info("Update: success for expense ID:", expenseID)
	c.SendSuccess(200, "Expense updated successfully", updated)
}

// Delete godoc
// @Title Delete Expense
// @Summary Delete an expense by ID
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses/:id [delete]
func (c *ExpenseController) Delete() {
	logs.Info("Delete expense endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	expenseID, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || expenseID <= 0 {
		logs.Warn("Delete: invalid expense ID param:", c.Ctx.Input.Param(":id"))
		c.SendError(400, "Invalid expense ID")
		return
	}

	// Confirm ownership before deleting
	existing, err := models.GetExpenseByID(expenseID, userID)
	if err != nil {
		logs.Error("Delete: error fetching expense:", err)
		c.SendError(500, "Failed to delete expense")
		return
	}
	if existing == nil {
		logs.Warn("Delete: expense not found or not owned, ID:", expenseID, "UserID:", userID)
		c.SendError(404, "Expense not found")
		return
	}

	if err := models.DeleteExpense(expenseID, userID); err != nil {
		logs.Error("Delete: failed to delete expense:", err)
		c.SendError(500, "Failed to delete expense")
		return
	}

	logs.Info("Delete: success for expense ID:", expenseID)
	c.SendSuccess(200, "Expense deleted successfully", nil)
}
