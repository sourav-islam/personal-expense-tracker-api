// Package models handles all data structures and CSV file operations.
package models

import (
	"sort"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
)

// FilterParams holds all supported query parameters for filtering
// and sorting an expense list.
type FilterParams struct {
	Category  string
	DateFrom  string
	DateTo    string
	SortBy    string
	SortOrder string
	Limit     int
}

// CategorySummary holds the spending total and count for one category.
type CategorySummary struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

// SummaryResult is the full response payload for the summary endpoint.
type SummaryResult struct {
	DateFrom    string            `json:"date_from"`
	DateTo      string            `json:"date_to"`
	TotalAmount float64           `json:"total_amount"`
	TotalCount  int               `json:"total_count"`
	ByCategory  []CategorySummary `json:"by_category"`
}

// ApplyFilters takes a slice of expenses and a FilterParams struct,
// then returns a filtered, sorted, and optionally limited slice.
func ApplyFilters(expenses []Expense, params FilterParams) []Expense {
	filtered := filterByCategory(expenses, params.Category)
	filtered = filterByDateRange(filtered, params.DateFrom, params.DateTo)
	filtered = sortExpenses(filtered, params.SortBy, params.SortOrder)

	if params.Limit > 0 && params.Limit < len(filtered) {
		filtered = filtered[:params.Limit]
	}

	return filtered
}

// filterByCategory returns only expenses matching the given category.
// If category is empty, the original slice is returned unchanged.
func filterByCategory(expenses []Expense, category string) []Expense {
	if category == "" {
		return expenses
	}

	var result []Expense
	for _, e := range expenses {
		if e.Category == category {
			result = append(result, e)
		}
	}

	logs.Info("filterByCategory: category=" + category +
		", matched: " + strconv.Itoa(len(result)) + " of " + strconv.Itoa(len(expenses)))
	return result
}

// filterByDateRange returns only expenses whose ExpenseDate falls within
// the given date range (inclusive on both ends).
// dateFrom and dateTo must be in YYYY-MM-DD format.
// Empty strings are treated as open-ended bounds.
func filterByDateRange(expenses []Expense, dateFrom, dateTo string) []Expense {
	if dateFrom == "" && dateTo == "" {
		return expenses
	}

	var result []Expense
	for _, e := range expenses {
		if dateFrom != "" && e.ExpenseDate < dateFrom {
			continue
		}
		if dateTo != "" && e.ExpenseDate > dateTo {
			continue
		}
		result = append(result, e)
	}

	logs.Info("filterByDateRange: from="+dateFrom+" to="+dateTo,
		"matched:", len(result), "of", len(expenses))
	return result
}

// sortExpenses sorts the given expense slice by the specified field and order.
// sortBy accepts "amount" or "expense_date" (default: "expense_date").
// sortOrder accepts "asc" or "desc" (default: "desc").
func sortExpenses(expenses []Expense, sortBy, sortOrder string) []Expense {
	if sortBy == "" {
		sortBy = "expense_date"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	ascending := sortOrder == "asc"

	sort.Slice(expenses, func(i, j int) bool {
		switch sortBy {
		case "amount":
			if ascending {
				return expenses[i].Amount < expenses[j].Amount
			}
			return expenses[i].Amount > expenses[j].Amount
		default:
			// Default: sort by expense_date
			// YYYY-MM-DD strings compare correctly as plain strings
			if ascending {
				return expenses[i].ExpenseDate < expenses[j].ExpenseDate
			}
			return expenses[i].ExpenseDate > expenses[j].ExpenseDate
		}
	})

	logs.Info("sortExpenses: sortBy=" + sortBy + " sortOrder=" + sortOrder)
	return expenses
}

// BuildSummary computes the total amount, total count, and per-category
// breakdown for the given slice of expenses.
func BuildSummary(expenses []Expense, dateFrom, dateTo string) SummaryResult {
	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)

	var totalAmount float64
	for _, e := range expenses {
		totalAmount += e.Amount
		categoryTotals[e.Category] += e.Amount
		categoryCounts[e.Category]++
	}

	// Build sorted category list for consistent output
	categoryKeys := make([]string, 0, len(categoryTotals))
	for k := range categoryTotals {
		categoryKeys = append(categoryKeys, k)
	}
	sort.Strings(categoryKeys)

	byCategory := make([]CategorySummary, 0, len(categoryKeys))
	for _, cat := range categoryKeys {
		byCategory = append(byCategory, CategorySummary{
			Category: cat,
			Total:    categoryTotals[cat],
			Count:    categoryCounts[cat],
		})
	}

	logs.Info("BuildSummary: total_amount=" + strconv.FormatFloat(totalAmount, 'f', -1, 64) +
		", total_count=" + strconv.Itoa(len(expenses)) +
		", categories=" + strconv.Itoa(len(byCategory)))

	return SummaryResult{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		TotalAmount: totalAmount,
		TotalCount:  len(expenses),
		ByCategory:  byCategory,
	}
}
