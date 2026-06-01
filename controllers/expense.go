package controllers

import (
	"strconv"

	"github.com/beego/beego/v2/core/logs"

	"expense-tracker-api/models"
)

// ExpenseController handles all expense CRUD, filtering, and summary endpoints.
type ExpenseController struct {
	BaseController
}

// expenseInput is the expected JSON body for creating or updating an expense.
type expenseInput struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
}


// Create a new expense
// Creates an expense record for the user identified by X-User-ID header

// @Success 201 {object} models.Expense "Expense created successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]interface{} "Unauthorized"

func (c *ExpenseController) Create() {
	userID, ok := c.GetAuthUserID()
	if !ok {
		return
	}

	var input expenseInput
	if err := c.ParseBody(&input); err != nil {
		logs.Warn("CreateExpense: failed to parse body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if err := models.ValidateExpenseInput(input.Title, input.Category, input.ExpenseDate, input.Amount); err != nil {
		c.SendError(400, err.Error())
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
		logs.Error("CreateExpense: failed to save:", err)
		c.SendError(500, "Internal server error")
		return
	}

	logs.Info("Expense created, id:", expense.ID, "user:", userID)
	c.SendCreated("Expense created successfully", expense)
}

/*List returns all expenses for the authenticated user with optional filters and pagination.

@Param X-User-ID header int true "Authenticated user ID"
@Param category query string false "Filter by category"
@Param date_from query string false "Start date (YYYY-MM-DD)"
@Param date_to query string false "End date (YYYY-MM-DD)"
@Param sort_by query string false "Sort field: amount or expense_date"
@Param sort_order query string false "Sort direction: asc or desc (default: desc)"
@Param limit query int false "Items per page (default: 10)"
@Param page query int false "Page number (default: 1)"
@Success 200 {object} map[string]interface{} "Expenses retrieved"
@Failure 401 {object} map[string]interface{} "Unauthorized"


*/


func (c *ExpenseController) List() {
	userID, ok := c.GetAuthUserID()
	if !ok {
		return
	}

	category  := c.GetString("category")
	dateFrom  := c.GetString("date_from")
	dateTo    := c.GetString("date_to")
	sortBy    := c.GetString("sort_by")
	sortOrder := c.GetString("sort_order")
	limitStr  := c.GetString("limit")
	pageStr   := c.GetString("page")

	if sortOrder == "" {
		sortOrder = "desc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		c.SendError(400, "sort_order must be 'asc' or 'desc'")
		return
	}
	if sortBy != "" && sortBy != "amount" && sortBy != "expense_date" {
		c.SendError(400, "sort_by must be 'amount' or 'expense_date'")
		return
	}

	expenses, err := models.GetExpensesByUserID(userID)
	if err != nil {
		logs.Error("ListExpenses: failed to read:", err)
		c.SendError(500, "Internal server error")
		return
	}

	expenses = models.FilterExpenses(expenses, category, dateFrom, dateTo, sortBy, sortOrder)

	limit := 10
	page  := 1
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	total := len(expenses)
	start := (page - 1) * limit
	end   := start + limit
	if start >= total {
		expenses = []models.Expense{}
	} else {
		if end > total {
			end = total
		}
		expenses = expenses[start:end]
	}

	c.SendSuccess("Expenses retrieved", map[string]interface{}{
		"expenses":    expenses,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	})
}

/*
GetOne returns a single expense by ID for the authenticated user.
@Description Returns the expense matching the given ID, owned by the authenticated user
@Param X-User-ID header int true "Authenticated user ID"
@Param id path int true "Expense ID"
@Success 200 {object} models.Expense "Expense retrieved"
@Failure 400 {object} map[string]interface{} "Invalid ID"
@Failure 401 {object} map[string]interface{} "Unauthorized"
@Failure 404 {object} map[string]interface{} "Expense not found"
@router /expenses/:id [get]
*/


func (c *ExpenseController) GetOne() {
	userID, ok := c.GetAuthUserID()
	if !ok {
		return
	}

	id, valid := c.parseExpenseID()
	if !valid {
		return
	}

	expense, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("GetExpense: failed to read:", err)
		c.SendError(500, "Internal server error")
		return
	}
	if expense == nil {
		c.SendError(404, "Expense not found")
		return
	}

	c.SendSuccess("Expense retrieved", expense)
}

/*
Update modifies an existing expense owned by the authenticated user.
Update an expense
@Description Replaces all fields of the expense with the given ID
@Param X-User-ID header int true "Authenticated user ID"
@Param id path int true "Expense ID"
@Param body body controllers.expenseInput true "Updated expense details"
@Success 200 {object} models.Expense "Expense updated successfully"
@Failure 400 {object} map[string]interface{} "Validation error"
@Failure 401 {object} map[string]interface{} "Unauthorized"
@Failure 404 {object} map[string]interface{} "Expense not found"
@router /expenses/:id [put]
*/

func (c *ExpenseController) Update() {
	userID, ok := c.GetAuthUserID()
	if !ok {
		return
	}

	id, valid := c.parseExpenseID()
	if !valid {
		return
	}

	existing, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("UpdateExpense: failed to fetch:", err)
		c.SendError(500, "Internal server error")
		return
	}
	if existing == nil {
		c.SendError(404, "Expense not found")
		return
	}

	var input expenseInput
	if err := c.ParseBody(&input); err != nil {
		logs.Warn("UpdateExpense: failed to parse body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if err := models.ValidateExpenseInput(input.Title, input.Category, input.ExpenseDate, input.Amount); err != nil {
		c.SendError(400, err.Error())
		return
	}

	existing.Title       = input.Title
	existing.Amount      = input.Amount
	existing.Category    = input.Category
	existing.Note        = input.Note
	existing.ExpenseDate = input.ExpenseDate

	if err := models.UpdateExpense(existing); err != nil {
		logs.Error("UpdateExpense: failed to save:", err)
		c.SendError(500, "Internal server error")
		return
	}

	logs.Info("Expense updated, id:", id, "user:", userID)
	c.SendSuccess("Expense updated successfully", existing)
}

/*
Delete removes an expense owned by the authenticated user.

Delete an expense
@Description Permanently removes the expense with the given ID
@Param X-User-ID header int true "Authenticated user ID"
@Param id path int true "Expense ID"
@Success 200 {object} map[string]interface{} "Expense deleted successfully"
@Failure 400 {object} map[string]interface{} "Invalid ID"
@Failure 401 {object} map[string]interface{} "Unauthorized"
@Failure 404 {object} map[string]interface{} "Expense not found"
@router /expenses/:id [delete]

*/

func (c *ExpenseController) Delete() {
	userID, ok := c.GetAuthUserID()
	if !ok {
		return
	}

	id, valid := c.parseExpenseID()
	if !valid {
		return
	}

	existing, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("DeleteExpense: failed to fetch:", err)
		c.SendError(500, "Internal server error")
		return
	}
	if existing == nil {
		c.SendError(404, "Expense not found")
		return
	}

	if err := models.DeleteExpense(id, userID); err != nil {
		logs.Error("DeleteExpense: failed to delete:", err)
		c.SendError(500, "Internal server error")
		return
	}

	logs.Info("Expense deleted, id:", id, "user:", userID)
	c.SendSuccess("Expense deleted successfully", nil)
}

/*
Summary returns aggregated spending data for the authenticated user in a date range.
Get spending summary
@Description Returns total spending and per-category breakdown for the given date range
@Param X-User-ID header int true "Authenticated user ID"
@Param date_from query string true "Start date (YYYY-MM-DD)"
@Param date_to query string true "End date (YYYY-MM-DD)"
@Success 200 {object} models.ExpenseSummary "Summary generated"
@Failure 400 {object} map[string]interface{} "Missing or invalid date params"
@Failure 401 {object} map[string]interface{} "Unauthorized"
@router /expenses/summary [get]
*/

func (c *ExpenseController) Summary() {
	userID, ok := c.GetAuthUserID()
	if !ok {
		return
	}

	dateFrom := c.GetString("date_from")
	dateTo   := c.GetString("date_to")

	if dateFrom == "" || dateTo == "" {
		c.SendError(400, "date_from and date_to are required")
		return
	}
	if dateFrom > dateTo {
		c.SendError(400, "date_from must be before or equal to date_to")
		return
	}

	summary, err := models.BuildSummary(userID, dateFrom, dateTo)
	if err != nil {
		logs.Error("Summary: failed to build:", err)
		c.SendError(500, "Internal server error")
		return
	}

	c.SendSuccess("Summary generated", summary)
}

// parseExpenseID extracts and validates the :id path parameter.
// Returns the integer ID and true on success, or 0 and false after sending a 400 error.
func (c *ExpenseController) parseExpenseID() (int, bool) {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.SendError(400, "Invalid expense ID")
		return 0, false
	}
	return id, true
}
