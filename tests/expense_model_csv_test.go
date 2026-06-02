package tests

import (
	"os"
	"path/filepath"
	"testing"

	"expense-tracker-api/models"
)

// newTempExpenseCSV creates a temp CSV file with header for expense tests.
func newTempExpenseCSV(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "expenses.csv")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp expense CSV: %v", err)
	}
	f.WriteString("id,user_id,title,amount,category,note,expense_date,created_at\n")
	f.Close()
	models.SetExpensesCSVPath(path)
	return path, func() {
		models.SetExpensesCSVPath("")
		os.Remove(path)
	}
}

// seedExpenses appends raw CSV rows to the expense file.
func seedExpenses(t *testing.T, path string, rows string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("seedExpenses: %v", err)
	}
	defer f.Close()
	f.WriteString(rows)
}

// TestGetExpensesByUserID tests fetching expenses filtered by user.
func TestGetExpensesByUserID(t *testing.T) {
	tests := []struct {
		name      string
		seedData  string
		userID    int
		wantCount int
	}{
		{
			name:      "empty CSV returns empty slice",
			seedData:  "",
			userID:    1,
			wantCount: 0,
		},
		{
			name:      "returns only matching user expenses",
			seedData:  "1,1,Lunch,350.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n2,2,Bus,50.00,Transport,,2025-06-01,2025-06-01T00:00:00Z\n3,1,Dinner,500.00,Food,,2025-06-02,2025-06-02T00:00:00Z\n",
			userID:    1,
			wantCount: 2,
		},
		{
			name:      "user with no expenses",
			seedData:  "1,1,Lunch,350.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n",
			userID:    99,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempExpenseCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedExpenses(t, path, tt.seedData)
			}

			expenses, err := models.GetExpensesByUserID(tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(expenses) != tt.wantCount {
				t.Errorf("got %d expenses, want %d", len(expenses), tt.wantCount)
			}
		})
	}
}

// TestGetExpenseByID tests ownership-enforced lookup.
func TestGetExpenseByID(t *testing.T) {
	seedRow := "1,1,Lunch,350.00,Food,note,2025-06-01,2025-06-01T00:00:00Z\n"

	tests := []struct {
		name      string
		id        int
		userID    int
		wantFound bool
	}{
		{
			name:      "correct id and owner",
			id:        1,
			userID:    1,
			wantFound: true,
		},
		{
			name:      "correct id wrong owner",
			id:        1,
			userID:    2,
			wantFound: false,
		},
		{
			name:      "wrong id correct owner",
			id:        99,
			userID:    1,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempExpenseCSV(t)
			defer cleanup()
			seedExpenses(t, path, seedRow)

			expense, err := models.GetExpenseByID(tt.id, tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantFound && expense == nil {
				t.Error("expected expense but got nil")
			}
			if !tt.wantFound && expense != nil {
				t.Error("expected nil but got expense")
			}
		})
	}
}

// TestCreateExpense tests appending a new expense to the CSV.
func TestCreateExpense(t *testing.T) {
	tests := []struct {
		name       string
		seedData   string
		input      models.Expense
		wantID     int
	}{
		{
			name:   "create first expense gets id 1",
			input:  models.Expense{UserID: 1, Title: "Lunch", Amount: 350, Category: "Food", ExpenseDate: "2025-06-01"},
			wantID: 1,
		},
		{
			name:     "create expense after existing gets id 2",
			seedData: "1,1,Breakfast,100.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n",
			input:    models.Expense{UserID: 1, Title: "Lunch", Amount: 350, Category: "Food", ExpenseDate: "2025-06-01"},
			wantID:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempExpenseCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedExpenses(t, path, tt.seedData)
			}

			err := models.CreateExpense(&tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.input.ID != tt.wantID {
				t.Errorf("got ID %d, want %d", tt.input.ID, tt.wantID)
			}
			if tt.input.CreatedAt == "" {
				t.Error("CreatedAt should be set after create")
			}

			// Verify persistence
			found, _ := models.GetExpenseByID(tt.input.ID, tt.input.UserID)
			if found == nil {
				t.Error("expense not found in CSV after create")
			}
		})
	}
}

// TestUpdateExpense tests updating an existing expense.
func TestUpdateExpense(t *testing.T) {
	tests := []struct {
		name        string
		seedData    string
		update      models.Expense
		wantErr     bool
		wantTitle   string
	}{
		{
			name:      "successful update",
			seedData:  "1,1,Lunch,350.00,Food,old note,2025-06-01,2025-06-01T00:00:00Z\n",
			update:    models.Expense{ID: 1, UserID: 1, Title: "Updated Lunch", Amount: 400, Category: "Food", Note: "new note", ExpenseDate: "2025-06-01", CreatedAt: "2025-06-01T00:00:00Z"},
			wantErr:   false,
			wantTitle: "Updated Lunch",
		},
		{
			name:     "update non-existent expense returns error",
			seedData: "1,1,Lunch,350.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n",
			update:   models.Expense{ID: 99, UserID: 1, Title: "X", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"},
			wantErr:  true,
		},
		{
			name:     "update wrong owner returns error",
			seedData: "1,1,Lunch,350.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n",
			update:   models.Expense{ID: 1, UserID: 2, Title: "X", Amount: 100, Category: "Food", ExpenseDate: "2025-06-01"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempExpenseCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedExpenses(t, path, tt.seedData)
			}

			err := models.UpdateExpense(&tt.update)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the update was persisted
			found, _ := models.GetExpenseByID(tt.update.ID, tt.update.UserID)
			if found == nil {
				t.Fatal("expense not found after update")
			}
			if found.Title != tt.wantTitle {
				t.Errorf("got title %q, want %q", found.Title, tt.wantTitle)
			}
		})
	}
}

// TestDeleteExpense tests removing an expense from the CSV.
func TestDeleteExpense(t *testing.T) {
	tests := []struct {
		name          string
		seedData      string
		deleteID      int
		deleteUserID  int
		wantErr       bool
		wantRemaining int
	}{
		{
			name:          "successful delete",
			seedData:      "1,1,Lunch,350.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n2,1,Bus,50.00,Transport,,2025-06-01,2025-06-01T00:00:00Z\n",
			deleteID:      1,
			deleteUserID:  1,
			wantErr:       false,
			wantRemaining: 1,
		},
		{
			name:         "delete non-existent returns error",
			seedData:     "1,1,Lunch,350.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n",
			deleteID:     99,
			deleteUserID: 1,
			wantErr:      true,
		},
		{
			name:         "delete wrong owner returns error",
			seedData:     "1,1,Lunch,350.00,Food,,2025-06-01,2025-06-01T00:00:00Z\n",
			deleteID:     1,
			deleteUserID: 2,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempExpenseCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedExpenses(t, path, tt.seedData)
			}

			err := models.DeleteExpense(tt.deleteID, tt.deleteUserID)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			remaining, _ := models.GetExpensesByUserID(tt.deleteUserID)
			if len(remaining) != tt.wantRemaining {
				t.Errorf("got %d remaining expenses, want %d", len(remaining), tt.wantRemaining)
			}
		})
	}
}

// TestBuildSummary tests the spending summary aggregation.
func TestBuildSummary(t *testing.T) {
	tests := []struct {
		name            string
		seedData        string
		userID          int
		dateFrom        string
		dateTo          string
		wantTotal       float64
		wantCount       int
		wantCategories  int
	}{
		{
			name:           "summary with multiple categories",
			seedData:       "1,1,Lunch,350.00,Food,,2025-06-10,2025-06-10T00:00:00Z\n2,1,Bus,50.00,Transport,,2025-06-11,2025-06-11T00:00:00Z\n3,1,Dinner,500.00,Food,,2025-06-12,2025-06-12T00:00:00Z\n",
			userID:         1,
			dateFrom:       "2025-06-01",
			dateTo:         "2025-06-30",
			wantTotal:      900.00,
			wantCount:      3,
			wantCategories: 2,
		},
		{
			name:           "empty date range returns zero",
			seedData:       "1,1,Lunch,350.00,Food,,2025-06-10,2025-06-10T00:00:00Z\n",
			userID:         1,
			dateFrom:       "2025-07-01",
			dateTo:         "2025-07-31",
			wantTotal:      0,
			wantCount:      0,
			wantCategories: 0,
		},
		{
			name:           "other user data not included",
			seedData:       "1,1,Lunch,350.00,Food,,2025-06-10,2025-06-10T00:00:00Z\n2,2,Bus,999.00,Transport,,2025-06-10,2025-06-10T00:00:00Z\n",
			userID:         1,
			dateFrom:       "2025-06-01",
			dateTo:         "2025-06-30",
			wantTotal:      350.00,
			wantCount:      1,
			wantCategories: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempExpenseCSV(t)
			defer cleanup()
			if tt.seedData != "" {
				seedExpenses(t, path, tt.seedData)
			}

			summary, err := models.BuildSummary(tt.userID, tt.dateFrom, tt.dateTo)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if summary.TotalAmount != tt.wantTotal {
				t.Errorf("TotalAmount = %.2f, want %.2f", summary.TotalAmount, tt.wantTotal)
			}
			if summary.TotalCount != tt.wantCount {
				t.Errorf("TotalCount = %d, want %d", summary.TotalCount, tt.wantCount)
			}
			if len(summary.ByCategory) != tt.wantCategories {
				t.Errorf("ByCategory count = %d, want %d", len(summary.ByCategory), tt.wantCategories)
			}
		})
	}
}
