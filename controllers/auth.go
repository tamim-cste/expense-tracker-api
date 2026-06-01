// Package controllers implements HTTP handlers for the Expense Tracker API.
package controllers

import (
	"github.com/beego/beego/v2/core/logs"

	"expense-tracker-api/models"
)

// AuthController handles user registration and login endpoints.
type AuthController struct {
	BaseController
}

// registerInput is the expected JSON body for the register endpoint.
type registerInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginInput is the expected JSON body for the login endpoint.
type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}


// Description Creates a new user with name, email, and password
// Success 201 {object} map[string]interface{} "User registered successfully"
// Failure 400 {object} map[string]interface{} "Validation error"
// Failure 409 {object} map[string]interface{} "Email already exists"

func (c *AuthController) Register() {
	var input registerInput
	if err := c.ParseBody(&input); err != nil {
		logs.Warn("Register: failed to parse body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if err := models.ValidateUserInput(input.Name, input.Email, input.Password); err != nil {
		c.SendError(400, err.Error())
		return
	}

	existing, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Register: failed to read users:", err)
		c.SendError(500, "Internal server error")
		return
	}
	if existing != nil {
		c.SendError(409, "Email already exists")
		return
	}

	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}
	if err := models.CreateUser(user); err != nil {
		logs.Error("Register: failed to create user:", err)
		c.SendError(500, "Internal server error")
		return
	}

	logs.Info("New user registered:", input.Email)
	c.SendCreated("User registered successfully", nil)
}


// Log in with email and password
// Success 200 {object} map[string]interface{} "Login successful"
// Failure 400 {object} map[string]interface{} "Missing fields"
// Failure 401 {object} map[string]interface{} "Invalid credentials"
// Description Authenticates a user and returns their ID, name, and email on success
func (c *AuthController) Login() {
	var input loginInput
	if err := c.ParseBody(&input); err != nil {
		logs.Warn("Login: failed to parse body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if input.Email == "" || input.Password == "" {
		c.SendError(400, "Email and password are required")
		return
	}

	user, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Login: failed to read users:", err)
		c.SendError(500, "Internal server error")
		return
	}
	if user == nil || user.Password != input.Password {
		c.SendError(401, "Invalid email or password")
		return
	}

	logs.Info("User logged in:", input.Email)
	c.SendSuccess("Login successful", map[string]interface{}{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
	})
}
