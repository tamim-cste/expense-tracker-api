package models

import (
	"encoding/csv"
	"errors"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
)

// This slice defines the valid expense categories.
var AllowedCategories = []string{
	"Food", "Transport", "Housing", "Entertainment",
	"Shopping", "Healthcare", "Education", "Utilities", "Other",
}

// Expense structure represents a single expense record.
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

// CategorySummary structure holds aggregated data per category.
type CategorySummary struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

// ExpenseSummary is the response body for the summary endpoint.
type ExpenseSummary struct {
	DateFrom   string            `json:"date_from"`
	DateTo     string            `json:"date_to"`
	TotalAmount float64          `json:"total_amount"`
	TotalCount  int              `json:"total_count"`
	ByCategory  []CategorySummary `json:"by_category"`
}

// getExpensesCSVPath returns the configured CSV file path for expenses.
func getExpensesCSVPath() string {
	path, _ := web.AppConfig.String("expenses_csv_path")
	if path == "" {
		path = "data/expenses.csv"
	}
	return path
}

// ensureExpensesCSV creates the CSV file with headers if it doesn't exist.
func ensureExpensesCSV(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll("data", 0755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w := csv.NewWriter(f)
		if err := w.Write([]string{
			"id", "user_id", "title", "amount", "category", "note", "expense_date", "created_at",
		}); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	}
	return nil
}

// readAllExpenses reads every expense row from the CSV.
func readAllExpenses() ([]Expense, error) {
	path := getExpensesCSVPath()
	if err := ensureExpensesCSV(path); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var expenses []Expense
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 8 {
			continue
		}
		id, _ := strconv.Atoi(record[0])
		userID, _ := strconv.Atoi(record[1])
		amount, _ := strconv.ParseFloat(record[3], 64)
		expenses = append(expenses, Expense{
			ID:          id,
			UserID:      userID,
			Title:       record[2],
			Amount:      amount,
			Category:    record[4],
			Note:        record[5],
			ExpenseDate: record[6],
			CreatedAt:   record[7],
		})
	}
	return expenses, nil
}

// writeAllExpenses rewrites the entire CSV with given expenses.
func writeAllExpenses(expenses []Expense) error {
	path := getExpensesCSVPath()
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{
		"id", "user_id", "title", "amount", "category", "note", "expense_date", "created_at",
	}); err != nil {
		return err
	}

	for _, e := range expenses {
		row := []string{
			strconv.Itoa(e.ID),
			strconv.Itoa(e.UserID),
			e.Title,
			strconv.FormatFloat(e.Amount, 'f', 2, 64),
			e.Category,
			e.Note,
			e.ExpenseDate,
			e.CreatedAt,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// GetNextExpenseID returns the next available expense ID.
func GetNextExpenseID() int {
	expenses, err := readAllExpenses()
	if err != nil || len(expenses) == 0 {
		return 1
	}
	max := 0
	for _, e := range expenses {
		if e.ID > max {
			max = e.ID
		}
	}
	return max + 1
}

// GetExpensesByUserID returns all expenses belonging to a user.
func GetExpensesByUserID(userID int) ([]Expense, error) {
	all, err := readAllExpenses()
	if err != nil {
		return nil, err
	}
	var result []Expense
	for _, e := range all {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetExpenseByID finds a specific expense by ID and userID (ownership check).
func GetExpenseByID(id int, userID int) (*Expense, error) {
	all, err := readAllExpenses()
	if err != nil {
		return nil, err
	}
	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			return &e, nil
		}
	}
	return nil, nil
}

// CreateExpense appends a new expense to the CSV file.
func CreateExpense(expense *Expense) error {
	path := getExpensesCSVPath()
	if err := ensureExpensesCSV(path); err != nil {
		return err
	}

	expense.ID = GetNextExpenseID()
	expense.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	err = w.Write([]string{
		strconv.Itoa(expense.ID),
		strconv.Itoa(expense.UserID),
		expense.Title,
		strconv.FormatFloat(expense.Amount, 'f', 2, 64),
		expense.Category,
		expense.Note,
		expense.ExpenseDate,
		expense.CreatedAt,
	})
	if err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// UpdateExpense rewrites the CSV with the updated expense record.
func UpdateExpense(updated *Expense) error {
	all, err := readAllExpenses()
	if err != nil {
		return err
	}
	found := false
	for i, e := range all {
		if e.ID == updated.ID && e.UserID == updated.UserID {
			all[i] = *updated
			found = true
			break
		}
	}
	if !found {
		return errors.New("expense not found")
	}
	return writeAllExpenses(all)
}

// DeleteExpense removes an expense from the CSV by rewriting without it.
func DeleteExpense(id int, userID int) error {
	all, err := readAllExpenses()
	if err != nil {
		return err
	}
	filtered := make([]Expense, 0, len(all))
	found := false
	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			found = true
			continue
		}
		filtered = append(filtered, e)
	}
	if !found {
		return errors.New("expense not found")
	}
	return writeAllExpenses(filtered)
}

// FilterExpenses applies category, date range, sort filters to a slice.
func FilterExpenses(expenses []Expense, category, dateFrom, dateTo, sortBy, sortOrder string) []Expense {
	result := expenses

	// Filter by category
	if category != "" {
		var filtered []Expense
		for _, e := range result {
			if strings.EqualFold(e.Category, category) {
				filtered = append(filtered, e)
			}
		}
		result = filtered
	}

	// Filter by date range
	if dateFrom != "" {
		var filtered []Expense
		for _, e := range result {
			if e.ExpenseDate >= dateFrom {
				filtered = append(filtered, e)
			}
		}
		result = filtered
	}
	if dateTo != "" {
		var filtered []Expense
		for _, e := range result {
			if e.ExpenseDate <= dateTo {
				filtered = append(filtered, e)
			}
		}
		result = filtered
	}

	// Sort
	if sortBy == "amount" {
		sort.Slice(result, func(i, j int) bool {
			if sortOrder == "asc" {
				return result[i].Amount < result[j].Amount
			}
			return result[i].Amount > result[j].Amount
		})
	} else {
		// default: sort by expense_date
		sort.Slice(result, func(i, j int) bool {
			if sortOrder == "asc" {
				return result[i].ExpenseDate < result[j].ExpenseDate
			}
			return result[i].ExpenseDate > result[j].ExpenseDate
		})
	}

	return result
}

// BuildSummary generates a spending summary for a date range.
func BuildSummary(userID int, dateFrom, dateTo string) (*ExpenseSummary, error) {
	expenses, err := GetExpensesByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Filter date range
	filtered := FilterExpenses(expenses, "", dateFrom, dateTo, "", "")

	totalAmount := 0.0
	categoryMap := make(map[string]CategorySummary)

	for _, e := range filtered {
		totalAmount += e.Amount
		cs := categoryMap[e.Category]
		cs.Category = e.Category
		cs.Total += e.Amount
		cs.Count++
		categoryMap[e.Category] = cs
	}

	var byCategory []CategorySummary
	for _, cs := range categoryMap {
		byCategory = append(byCategory, cs)
	}
	sort.Slice(byCategory, func(i, j int) bool {
		return byCategory[i].Total > byCategory[j].Total
	})

	return &ExpenseSummary{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		TotalAmount: totalAmount,
		TotalCount:  len(filtered),
		ByCategory:  byCategory,
	}, nil
}

// IsValidCategory checks if a category string is in the allowed list.
func IsValidCategory(category string) bool {
	for _, c := range AllowedCategories {
		if c == category {
			return true
		}
	}
	return false
}

// ValidateExpenseInput validates expense creation/update fields.
func ValidateExpenseInput(title, category, expenseDate string, amount float64) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("Title is required")
	}
	if amount <= 0 {
		return errors.New("Amount must be a positive number")
	}
	if strings.TrimSpace(expenseDate) == "" {
		return errors.New("expense_date is required")
	}
	// Validate YYYY-MM-DD format
	_, err := time.Parse("2006-01-02", expenseDate)
	if err != nil {
		return errors.New("expense_date must be in YYYY-MM-DD format")
	}
	if strings.TrimSpace(category) == "" {
		return errors.New("Category is required")
	}
	if !IsValidCategory(category) {
		return errors.New("Invalid category. Allowed: Food, Transport, Housing, Entertainment, Shopping, Healthcare, Education, Utilities, Other")
	}
	return nil
}
