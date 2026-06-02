package models

import (
	"testing"
)

// buildExpenses is a helper to create a slice of Expense values for filter tests.
func buildExpenses() []Expense {
	return []Expense{
		{ID: 1, UserID: 1, Title: "Lunch", Amount: 350.50,
			Category: "Food", ExpenseDate: "2025-06-01"},
		{ID: 2, UserID: 1, Title: "Bus", Amount: 120.00,
			Category: "Transport", ExpenseDate: "2025-06-03"},
		{ID: 3, UserID: 1, Title: "Groceries", Amount: 800.00,
			Category: "Food", ExpenseDate: "2025-06-05"},
		{ID: 4, UserID: 1, Title: "Netflix", Amount: 650.00,
			Category: "Entertainment", ExpenseDate: "2025-06-10"},
		{ID: 5, UserID: 1, Title: "Electricity", Amount: 1200.00,
			Category: "Utilities", ExpenseDate: "2025-05-15"},
	}
}

// ---------------------------------------------------------------------------
// ApplyFilters — category
// ---------------------------------------------------------------------------

func TestApplyFilters_Category(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		wantCount int
	}{
		{name: "filter Food", category: "Food", wantCount: 2},
		{name: "filter Transport", category: "Transport", wantCount: 1},
		{name: "filter Utilities", category: "Utilities", wantCount: 1},
		{name: "no match", category: "Healthcare", wantCount: 0},
		{name: "empty = all", category: "", wantCount: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(buildExpenses(), FilterParams{Category: tt.category})
			if len(result) != tt.wantCount {
				t.Errorf("category=%q: got %d results, want %d",
					tt.category, len(result), tt.wantCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ApplyFilters — date range
// ---------------------------------------------------------------------------

func TestApplyFilters_DateRange(t *testing.T) {
	tests := []struct {
		name      string
		dateFrom  string
		dateTo    string
		wantCount int
	}{
		{name: "full June", dateFrom: "2025-06-01", dateTo: "2025-06-30", wantCount: 4},
		{name: "May only", dateFrom: "2025-05-01", dateTo: "2025-05-31", wantCount: 1},
		{name: "single day match", dateFrom: "2025-06-03", dateTo: "2025-06-03", wantCount: 1},
		{name: "single day no match", dateFrom: "2025-06-04", dateTo: "2025-06-04", wantCount: 0},
		{name: "only date_from", dateFrom: "2025-06-05", dateTo: "", wantCount: 2},
		{name: "only date_to", dateFrom: "", dateTo: "2025-05-31", wantCount: 1},
		{name: "no range = all", dateFrom: "", dateTo: "", wantCount: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(buildExpenses(), FilterParams{
				DateFrom: tt.dateFrom,
				DateTo:   tt.dateTo,
			})
			if len(result) != tt.wantCount {
				t.Errorf("from=%q to=%q: got %d results, want %d",
					tt.dateFrom, tt.dateTo, len(result), tt.wantCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ApplyFilters — sorting
// ---------------------------------------------------------------------------

func TestApplyFilters_SortByAmount(t *testing.T) {
	tests := []struct {
		name        string
		sortOrder   string
		wantFirstID int // ID of the expense that should come first
	}{
		{name: "amount desc (highest first)", sortOrder: "desc", wantFirstID: 5},
		{name: "amount asc (lowest first)", sortOrder: "asc", wantFirstID: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(buildExpenses(), FilterParams{
				SortBy:    "amount",
				SortOrder: tt.sortOrder,
			})
			if len(result) == 0 {
				t.Fatal("expected non-empty result")
			}
			if result[0].ID != tt.wantFirstID {
				t.Errorf("sort order=%q: first ID=%d, want %d",
					tt.sortOrder, result[0].ID, tt.wantFirstID)
			}
		})
	}
}

func TestApplyFilters_SortByDate(t *testing.T) {
	tests := []struct {
		name        string
		sortOrder   string
		wantFirstID int
	}{
		{name: "date desc (newest first)", sortOrder: "desc", wantFirstID: 4},
		{name: "date asc (oldest first)", sortOrder: "asc", wantFirstID: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(buildExpenses(), FilterParams{
				SortBy:    "expense_date",
				SortOrder: tt.sortOrder,
			})
			if len(result) == 0 {
				t.Fatal("expected non-empty result")
			}
			if result[0].ID != tt.wantFirstID {
				t.Errorf("sort order=%q: first ID=%d, want %d",
					tt.sortOrder, result[0].ID, tt.wantFirstID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ApplyFilters — limit
// ---------------------------------------------------------------------------

func TestApplyFilters_Limit(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{name: "limit 2", limit: 2, wantCount: 2},
		{name: "limit 10 (more than total)", limit: 10, wantCount: 5},
		{name: "limit 0 (disabled)", limit: 0, wantCount: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(buildExpenses(), FilterParams{Limit: tt.limit})
			if len(result) != tt.wantCount {
				t.Errorf("limit=%d: got %d results, want %d",
					tt.limit, len(result), tt.wantCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ApplyFilters — combined
// ---------------------------------------------------------------------------

func TestApplyFilters_Combined(t *testing.T) {
	// Food in June, cheapest first, max 1
	result := ApplyFilters(buildExpenses(), FilterParams{
		Category:  "Food",
		DateFrom:  "2025-06-01",
		DateTo:    "2025-06-30",
		SortBy:    "amount",
		SortOrder: "asc",
		Limit:     1,
	})
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("expected cheapest Food = ID 1, got ID %d", result[0].ID)
	}
}

// ---------------------------------------------------------------------------
// BuildSummary
// ---------------------------------------------------------------------------

func TestBuildSummary_TotalsAndCounts(t *testing.T) {
	expenses := []Expense{
		{Category: "Food", Amount: 350.50},
		{Category: "Food", Amount: 800.00},
		{Category: "Transport", Amount: 120.00},
		{Category: "Entertainment", Amount: 650.00},
	}

	result := BuildSummary(expenses, "2025-06-01", "2025-06-30")

	expectedTotal := 350.50 + 800.00 + 120.00 + 650.00
	if result.TotalAmount != expectedTotal {
		t.Errorf("TotalAmount=%.2f, want %.2f", result.TotalAmount, expectedTotal)
	}
	if result.TotalCount != 4 {
		t.Errorf("TotalCount=%d, want 4", result.TotalCount)
	}
	if result.DateFrom != "2025-06-01" {
		t.Errorf("DateFrom=%q, want 2025-06-01", result.DateFrom)
	}
	if result.DateTo != "2025-06-30" {
		t.Errorf("DateTo=%q, want 2025-06-30", result.DateTo)
	}
	if len(result.ByCategory) != 3 {
		t.Errorf("ByCategory len=%d, want 3", len(result.ByCategory))
	}
}

func TestBuildSummary_Empty(t *testing.T) {
	result := BuildSummary([]Expense{}, "2025-06-01", "2025-06-30")
	if result.TotalAmount != 0 {
		t.Errorf("expected TotalAmount=0, got %.2f", result.TotalAmount)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected TotalCount=0, got %d", result.TotalCount)
	}
	if len(result.ByCategory) != 0 {
		t.Errorf("expected empty ByCategory, got %d", len(result.ByCategory))
	}
}

func TestBuildSummary_CategorySorted(t *testing.T) {
	// Categories should be alphabetically sorted in output
	expenses := []Expense{
		{Category: "Utilities", Amount: 100},
		{Category: "Food", Amount: 200},
		{Category: "Entertainment", Amount: 300},
	}
	result := BuildSummary(expenses, "2025-06-01", "2025-06-30")

	order := []string{"Entertainment", "Food", "Utilities"}
	for i, cat := range order {
		if result.ByCategory[i].Category != cat {
			t.Errorf("ByCategory[%d]=%q, want %q", i, result.ByCategory[i].Category, cat)
		}
	}
}
