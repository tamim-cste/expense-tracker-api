package models_test

import (
	"testing"

	"expense-tracker-api/models"
)

// TestValidateExpenseInput tests expense creation/update validation.
func TestValidateExpenseInput(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		category    string
		expenseDate string
		amount      float64
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid expense",
			title:       "Lunch",
			category:    "Food",
			expenseDate: "2025-06-10",
			amount:      350.50,
			wantErr:     false,
		},
		{
			name:        "missing title",
			title:       "",
			category:    "Food",
			expenseDate: "2025-06-10",
			amount:      350.50,
			wantErr:     true,
			errMsg:      "Title is required",
		},
		{
			name:        "zero amount",
			title:       "Lunch",
			category:    "Food",
			expenseDate: "2025-06-10",
			amount:      0,
			wantErr:     true,
			errMsg:      "Amount must be a positive number",
		},
		{
			name:        "negative amount",
			title:       "Lunch",
			category:    "Food",
			expenseDate: "2025-06-10",
			amount:      -100,
			wantErr:     true,
			errMsg:      "Amount must be a positive number",
		},
		{
			name:        "missing expense_date",
			title:       "Lunch",
			category:    "Food",
			expenseDate: "",
			amount:      100,
			wantErr:     true,
			errMsg:      "expense_date is required",
		},
		{
			name:        "invalid date format",
			title:       "Lunch",
			category:    "Food",
			expenseDate: "10-06-2025",
			amount:      100,
			wantErr:     true,
			errMsg:      "expense_date must be in YYYY-MM-DD format",
		},
		{
			name:        "missing category",
			title:       "Lunch",
			category:    "",
			expenseDate: "2025-06-10",
			amount:      100,
			wantErr:     true,
			errMsg:      "Category is required",
		},
		{
			name:        "invalid category",
			title:       "Lunch",
			category:    "InvalidCat",
			expenseDate: "2025-06-10",
			amount:      100,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateExpenseInput(tt.title, tt.category, tt.expenseDate, tt.amount)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestIsValidCategory tests category validation.
func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		category string
		valid    bool
	}{
		{"Food", true},
		{"Transport", true},
		{"Housing", true},
		{"Entertainment", true},
		{"Shopping", true},
		{"Healthcare", true},
		{"Education", true},
		{"Utilities", true},
		{"Other", true},
		{"food", false}, // case sensitive
		{"FOOD", false},
		{"Invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := models.IsValidCategory(tt.category)
			if got != tt.valid {
				t.Errorf("IsValidCategory(%q) = %v, want %v", tt.category, got, tt.valid)
			}
		})
	}
}

// TestFilterExpenses tests the filtering and sorting logic.
func TestFilterExpenses(t *testing.T) {
	expenses := []models.Expense{
		{ID: 1, UserID: 1, Title: "Lunch", Amount: 350, Category: "Food", ExpenseDate: "2025-06-10"},
		{ID: 2, UserID: 1, Title: "Bus", Amount: 50, Category: "Transport", ExpenseDate: "2025-06-11"},
		{ID: 3, UserID: 1, Title: "Dinner", Amount: 500, Category: "Food", ExpenseDate: "2025-06-12"},
		{ID: 4, UserID: 1, Title: "Movie", Amount: 200, Category: "Entertainment", ExpenseDate: "2025-06-15"},
	}

	t.Run("filter by category", func(t *testing.T) {
		result := models.FilterExpenses(expenses, "Food", "", "", "", "desc")
		if len(result) != 2 {
			t.Errorf("expected 2 food expenses, got %d", len(result))
		}
	})

	t.Run("filter by date range", func(t *testing.T) {
		result := models.FilterExpenses(expenses, "", "2025-06-11", "2025-06-12", "", "desc")
		if len(result) != 2 {
			t.Errorf("expected 2 expenses in range, got %d", len(result))
		}
	})

	t.Run("sort by amount asc", func(t *testing.T) {
		result := models.FilterExpenses(expenses, "", "", "", "amount", "asc")
		if result[0].Amount > result[1].Amount {
			t.Error("expected ascending order by amount")
		}
	})

	t.Run("sort by amount desc", func(t *testing.T) {
		result := models.FilterExpenses(expenses, "", "", "", "amount", "desc")
		if result[0].Amount < result[1].Amount {
			t.Error("expected descending order by amount")
		}
	})

	t.Run("no filter returns all", func(t *testing.T) {
		result := models.FilterExpenses(expenses, "", "", "", "", "desc")
		if len(result) != 4 {
			t.Errorf("expected 4 expenses, got %d", len(result))
		}
	})
}
