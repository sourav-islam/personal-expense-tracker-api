// Package controllers handles all incoming HTTP requests.
package controllers

import (
	"encoding/json"
	"expense-tracker-api/models"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
)

// ExpenseController handles all expense CRUD, filtering, and summary operations.
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

// validateExpenseInput checks all required and optional fields for
// create and update operations.
// Returns a non-empty error message string if validation fails.
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

// parseFilterParams reads and validates all filter/sort query parameters
// from the request. Returns a FilterParams struct and an error message
// string (empty if all params are valid).
func (c *ExpenseController) parseFilterParams() (models.FilterParams, string) {
	params := models.FilterParams{}

	// --- category ---
	category := c.GetString("category")
	if category != "" && !models.IsValidCategory(category) {
		return params, "Invalid category filter value"
	}
	params.Category = category

	// --- date_from ---
	dateFrom := c.GetString("date_from")
	if dateFrom != "" && !models.IsValidDateFormat(dateFrom) {
		return params, "Invalid date_from format, expected YYYY-MM-DD"
	}
	params.DateFrom = dateFrom

	// --- date_to ---
	dateTo := c.GetString("date_to")
	if dateTo != "" && !models.IsValidDateFormat(dateTo) {
		return params, "Invalid date_to format, expected YYYY-MM-DD"
	}
	params.DateTo = dateTo

	// --- date range logic check ---
	if dateFrom != "" && dateTo != "" && dateFrom > dateTo {
		return params, "date_from must not be after date_to"
	}

	// --- sort_by ---
	sortBy := c.GetString("sort_by")
	if sortBy != "" && sortBy != "amount" && sortBy != "expense_date" {
		return params, "Invalid sort_by value, accepted: amount, expense_date"
	}
	params.SortBy = sortBy

	// --- sort_order ---
	sortOrder := c.GetString("sort_order")
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return params, "Invalid sort_order value, accepted: asc, desc"
	}
	params.SortOrder = sortOrder

	// --- limit ---
	limitStr := c.GetString("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return params, "Invalid limit parameter, must be a positive integer"
		}
		params.Limit = limit
	}

	return params, ""
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

	if errMsg := validateExpenseInput(
		input.Title, input.Category, input.ExpenseDate, input.Amount,
	); errMsg != "" {
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
		logs.Error("Create expense: failed to save:", err)
		c.SendError(500, "Failed to create expense")
		return
	}

	logs.Info("Create expense: success ID:", expense.ID)
	c.SendSuccess(201, "Expense created successfully", expense)
}

// List godoc
// @Title List Expenses
// @Summary List expenses for the authenticated user with optional filters and sorting
// @Param X-User-ID header int true "User ID"
// @Param category query string false "Filter by category"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param sort_by query string false "Sort field: amount or expense_date"
// @Param sort_order query string false "Sort direction: asc or desc"
// @Param limit query int false "Max number of results"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses [get]
func (c *ExpenseController) List() {
	logs.Info("List expenses endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	params, errMsg := c.parseFilterParams()
	if errMsg != "" {
		logs.Warn("List expenses: invalid query params:", errMsg)
		c.SendError(400, errMsg)
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("List expenses: failed to retrieve:", err)
		c.SendError(500, "Failed to retrieve expenses")
		return
	}

	expenses = models.ApplyFilters(expenses, params)

	// Always return an array, never null
	if expenses == nil {
		expenses = []models.Expense{}
	}

	logs.Info("List expenses: returning", len(expenses), "for userID:", userID)
	c.SendSuccess(200, "Expenses retrieved", expenses)
}

// GetOne godoc
// @Title Get Expense
// @Summary Get a single expense by ID
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses/:id [get]
func (c *ExpenseController) GetOne() {
	logs.Info("GetOne expense endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	expenseID, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil || expenseID <= 0 {
		logs.Warn("GetOne: invalid expense ID:", c.Ctx.Input.Param(":id"))
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
		logs.Warn("GetOne: not found, ID:", expenseID, "userID:", userID)
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
		logs.Warn("Update: invalid expense ID:", c.Ctx.Input.Param(":id"))
		c.SendError(400, "Invalid expense ID")
		return
	}

	existing, err := models.GetExpenseByID(expenseID, userID)
	if err != nil {
		logs.Error("Update: error fetching expense:", err)
		c.SendError(500, "Failed to update expense")
		return
	}
	if existing == nil {
		logs.Warn("Update: not found or not owned, ID:", expenseID, "userID:", userID)
		c.SendError(404, "Expense not found")
		return
	}

	var input updateExpenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Update: failed to parse body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if errMsg := validateExpenseInput(
		input.Title, input.Category, input.ExpenseDate, input.Amount,
	); errMsg != "" {
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
		logs.Error("Update: failed to update:", err)
		c.SendError(500, "Failed to update expense")
		return
	}

	logs.Info("Update: success, ID:", expenseID)
	c.SendSuccess(200, "Expense updated successfully", updated)
}

// Delete godoc
// @Title Delete Expense
// @Summary Delete an expense by ID
// @Param X-User-ID header int true "User ID"
// @Param id path int true "Expense ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
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
		logs.Warn("Delete: invalid expense ID:", c.Ctx.Input.Param(":id"))
		c.SendError(400, "Invalid expense ID")
		return
	}

	existing, err := models.GetExpenseByID(expenseID, userID)
	if err != nil {
		logs.Error("Delete: error fetching expense:", err)
		c.SendError(500, "Failed to delete expense")
		return
	}
	if existing == nil {
		logs.Warn("Delete: not found or not owned, ID:", expenseID, "userID:", userID)
		c.SendError(404, "Expense not found")
		return
	}

	if err := models.DeleteExpense(expenseID, userID); err != nil {
		logs.Error("Delete: failed:", err)
		c.SendError(500, "Failed to delete expense")
		return
	}

	logs.Info("Delete: success, ID:", expenseID)
	c.SendSuccess(200, "Expense deleted successfully", nil)
}

// Summary godoc
// @Title Expense Summary
// @Summary Get a spending summary grouped by category for the authenticated user
// @Param X-User-ID header int true "User ID"
// @Param date_from query string true "Start date (YYYY-MM-DD)"
// @Param date_to query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/expenses/summary [get]
func (c *ExpenseController) Summary() {
	logs.Info("Summary endpoint called")

	userID := AuthMiddleware(&c.BaseController)
	if userID == 0 {
		return
	}

	// Both date params are required for summary
	dateFrom := c.GetString("date_from")
	dateTo := c.GetString("date_to")

	if dateFrom == "" {
		c.SendError(400, "date_from is required")
		return
	}
	if dateTo == "" {
		c.SendError(400, "date_to is required")
		return
	}
	if !models.IsValidDateFormat(dateFrom) {
		c.SendError(400, "Invalid date_from format, expected YYYY-MM-DD")
		return
	}
	if !models.IsValidDateFormat(dateTo) {
		c.SendError(400, "Invalid date_to format, expected YYYY-MM-DD")
		return
	}
	if dateFrom > dateTo {
		c.SendError(400, "date_from must not be after date_to")
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("Summary: failed to retrieve expenses:", err)
		c.SendError(500, "Failed to generate summary")
		return
	}

	// Filter to the requested date range only
	filtered := models.ApplyFilters(expenses, models.FilterParams{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	})

	result := models.BuildSummary(filtered, dateFrom, dateTo)

	logs.Info("Summary: generated for userID:", userID,
		"total:", result.TotalAmount, "count:", result.TotalCount)
	c.SendSuccess(200, "Summary generated", result)
}
