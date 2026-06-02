package controllers_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"expense-tracker-api/controllers"
	"expense-tracker-api/models"
)

func newTempExpenseCSV(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "expenses.csv")
	if err := os.WriteFile(path, []byte("id,user_id,title,amount,category,note,expense_date,created_at\n"), 0644); err != nil {
		t.Fatalf("failed to create temp expense CSV: %v", err)
	}
	return path, func() {
		models.SetExpensesCSVPath("")
		os.Remove(path)
	}
}

func newTempUserCSVForExpense(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "users.csv")
	if err := os.WriteFile(path, []byte("id,name,email,password,created_at\n"), 0644); err != nil {
		t.Fatalf("failed to create temp user CSV: %v", err)
	}
	return path, func() {
		models.SetUsersCSVPath("")
		os.Remove(path)
	}
}

func seedUsers(t *testing.T, path string, rows string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("id,name,email,password,created_at\n"+rows), 0644); err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}
}

func seedExpenses(t *testing.T, path string, rows string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("id,user_id,title,amount,category,note,expense_date,created_at\n"+rows), 0644); err != nil {
		t.Fatalf("failed to seed expenses: %v", err)
	}
}

func TestExpenseCreate(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "valid expense",
			body:        `{"title":"Lunch","amount":350.50,"category":"Food","note":"Team lunch","expense_date":"2025-06-10"}`,
			wantStatus:  201,
			wantMessage: "Expense created successfully",
		},
		{
			name:        "missing title",
			body:        `{"amount":350.50,"category":"Food","note":"Team lunch","expense_date":"2025-06-10"}`,
			wantStatus:  400,
			wantMessage: "Title is required",
		},
		{
			name:        "invalid category",
			body:        `{"title":"Lunch","amount":350.50,"category":"InvalidCat","note":"Team lunch","expense_date":"2025-06-10"}`,
			wantStatus:  400,
			wantMessage: "Invalid category. Allowed: Food, Transport, Housing, Entertainment, Shopping, Healthcare, Education, Utilities, Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userPath, userCleanup := newTempUserCSVForExpense(t)
			defer userCleanup()
			models.SetUsersCSVPath(userPath)
			seedUsers(t, userPath, "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n")

			expensePath, expenseCleanup := newTempExpenseCSV(t)
			defer expenseCleanup()
			models.SetExpensesCSVPath(expensePath)

			ctx, rec := buildControllerContext(t, http.MethodPost, "/api/v1/expenses", []byte(tt.body), map[string]string{
				"Content-Type": "application/json",
				"X-User-ID":    "1",
			})
			controller := controllers.ExpenseController{}
			controller.Ctx = ctx
			controller.Data = make(map[interface{}]interface{})
			controller.Prepare()
			executeControllerAction(t, controller.Create)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp apiResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.Message != tt.wantMessage {
				t.Fatalf("expected message %q, got %q", tt.wantMessage, resp.Message)
			}

			if tt.wantStatus == 201 {
				if resp.Data == nil {
					t.Fatal("expected returned expense data")
				}
				if resp.Data["title"] != "Lunch" {
					t.Fatalf("expected title Lunch, got %v", resp.Data["title"])
				}
			}
		})
	}
}

func TestExpenseSummaryAndList(t *testing.T) {
	userPath, userCleanup := newTempUserCSVForExpense(t)
	defer userCleanup()
	models.SetUsersCSVPath(userPath)
	seedUsers(t, userPath, "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n")

	expensePath, expenseCleanup := newTempExpenseCSV(t)
	defer expenseCleanup()
	models.SetExpensesCSVPath(expensePath)
	seedExpenses(t, expensePath, "1,1,Lunch,350.50,Food,Team lunch,2025-06-10,2025-06-10T14:30:00Z\n2,1,Bus,50.00,Transport,,2025-06-12,2025-06-12T09:00:00Z\n")

	// Verify list endpoint
	ctxList, recList := buildControllerContext(t, http.MethodGet, "/api/v1/expenses?limit=10&page=1", nil, map[string]string{
		"X-User-ID": "1",
	})
	listController := controllers.ExpenseController{}
	listController.Ctx = ctxList
	listController.Data = make(map[interface{}]interface{})
	listController.Prepare()
	executeControllerAction(t, listController.List)

	if recList.Code != 200 {
		t.Fatalf("expected list status 200, got %d", recList.Code)
	}

	var listResp apiResponse
	if err := json.Unmarshal(recList.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if listResp.Data == nil {
		t.Fatal("expected list response data")
	}

	// Verify summary endpoint
	summaryURL := "/api/v1/expenses/summary?date_from=2025-06-01&date_to=2025-06-30"
	ctxSummary, recSummary := buildControllerContext(t, http.MethodGet, summaryURL, nil, map[string]string{
		"X-User-ID": "1",
	})
	summaryController := controllers.ExpenseController{}
	summaryController.Ctx = ctxSummary
	summaryController.Data = make(map[interface{}]interface{})
	summaryController.Prepare()
	executeControllerAction(t, summaryController.Summary)

	if recSummary.Code != 200 {
		t.Fatalf("expected summary status 200, got %d", recSummary.Code)
	}

	var summaryResp apiResponse
	if err := json.Unmarshal(recSummary.Body.Bytes(), &summaryResp); err != nil {
		t.Fatalf("failed to parse summary response: %v", err)
	}
	if summaryResp.Data == nil {
		t.Fatal("expected summary response data")
	}
	if summaryResp.Data["total_count"].(float64) != 2 {
		t.Fatalf("expected total_count 2, got %v", summaryResp.Data["total_count"])
	}
}

func TestExpenseUnauthorized(t *testing.T) {
	expensePath, expenseCleanup := newTempExpenseCSV(t)
	defer expenseCleanup()
	models.SetExpensesCSVPath(expensePath)

	ctx, rec := buildControllerContext(t, http.MethodPost, "/api/v1/expenses", []byte(`{"title":"Lunch","amount":350.50,"category":"Food","note":"Team lunch","expense_date":"2025-06-10"}`), map[string]string{
		"Content-Type": "application/json",
	})

	controller := controllers.ExpenseController{}
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})
	controller.Prepare()
	executeControllerAction(t, controller.Create)

	if rec.Code != 401 {
		t.Fatalf("expected unauthorized status 401, got %d", rec.Code)
	}

	var resp apiResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Message != "Unauthorized" {
		t.Fatalf("expected message Unauthorized, got %q", resp.Message)
	}
}

func TestExpenseListFilters(t *testing.T) {
	userPath, userCleanup := newTempUserCSVForExpense(t)
	defer userCleanup()
	models.SetUsersCSVPath(userPath)
	seedUsers(t, userPath, "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n")

	expensePath, expenseCleanup := newTempExpenseCSV(t)
	defer expenseCleanup()
	models.SetExpensesCSVPath(expensePath)
	seedExpenses(t, expensePath, "1,1,Lunch,350.50,Food,Team lunch,2025-06-10,2025-06-10T14:30:00Z\n2,1,Bus,50.00,Transport,,2025-06-12,2025-06-12T09:00:00Z\n3,1,Dinner,500.00,Food,,2025-06-15,2025-06-15T19:00:00Z\n")

	tests := []struct {
		name     string
		url      string
		wantCode int
		checkLen bool
		wantLen  int
	}{
		{
			name:     "filter by category",
			url:      "/api/v1/expenses?category=Food",
			wantCode: 200,
			checkLen: true,
			wantLen:  2,
		},
		{
			name:     "filter by date range",
			url:      "/api/v1/expenses?date_from=2025-06-10&date_to=2025-06-12",
			wantCode: 200,
			checkLen: true,
			wantLen:  2,
		},
		{
			name:     "sort by amount desc",
			url:      "/api/v1/expenses?sort_by=amount&sort_order=desc",
			wantCode: 200,
			checkLen: true,
			wantLen:  3,
		},
		{
			name:     "invalid sort_order",
			url:      "/api/v1/expenses?sort_order=invalid",
			wantCode: 400,
		},
		{
			name:     "invalid sort_by",
			url:      "/api/v1/expenses?sort_by=invalid",
			wantCode: 400,
		},
		{
			name:     "pagination page 2",
			url:      "/api/v1/expenses?limit=1&page=2",
			wantCode: 200,
			checkLen: true,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := buildControllerContext(t, http.MethodGet, tt.url, nil, map[string]string{
				"X-User-ID": "1",
			})
			controller := controllers.ExpenseController{}
			controller.Ctx = ctx
			controller.Data = make(map[interface{}]interface{})
			controller.Prepare()
			executeControllerAction(t, controller.List)

			if rec.Code != tt.wantCode {
				t.Fatalf("expected status %d, got %d", tt.wantCode, rec.Code)
			}

			if tt.checkLen {
				var resp apiResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected response data")
				}
				expenses := resp.Data["expenses"].([]interface{})
				if len(expenses) != tt.wantLen {
					t.Fatalf("expected %d expenses, got %d", tt.wantLen, len(expenses))
				}
			}
		})
	}
}

func TestExpenseSummaryEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "missing date_from",
			url:         "/api/v1/expenses/summary?date_to=2025-06-30",
			wantStatus:  400,
			wantMessage: "date_from and date_to are required",
		},
		{
			name:        "missing date_to",
			url:         "/api/v1/expenses/summary?date_from=2025-06-01",
			wantStatus:  400,
			wantMessage: "date_from and date_to are required",
		},
		{
			name:        "invalid date range",
			url:         "/api/v1/expenses/summary?date_from=2025-06-30&date_to=2025-06-01",
			wantStatus:  400,
			wantMessage: "date_from must be before or equal to date_to",
		},
	}

	userPath, userCleanup := newTempUserCSVForExpense(t)
	defer userCleanup()
	models.SetUsersCSVPath(userPath)
	seedUsers(t, userPath, "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n")

	expensePath, expenseCleanup := newTempExpenseCSV(t)
	defer expenseCleanup()
	models.SetExpensesCSVPath(expensePath)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := buildControllerContext(t, http.MethodGet, tt.url, nil, map[string]string{
				"X-User-ID": "1",
			})
			controller := controllers.ExpenseController{}
			controller.Ctx = ctx
			controller.Data = make(map[interface{}]interface{})
			controller.Prepare()
			executeControllerAction(t, controller.Summary)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp apiResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.Message != tt.wantMessage {
				t.Fatalf("expected message %q, got %q", tt.wantMessage, resp.Message)
			}
		})
	}
}
