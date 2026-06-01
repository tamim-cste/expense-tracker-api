package controllers

import (
	"encoding/json"
	"strconv"

	"github.com/beego/beego/v2/server/web"

	"expense-tracker-api/models"
)

// BaseController provides shared utilities for all controllers.
type BaseController struct {
	web.Controller
}

// Prepare runs before every action — disables auto-render globally.
func (c *BaseController) Prepare() {
	c.EnableRender = false
}

// response is the standard API response envelope.
type response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendSuccess sends a 200 OK JSON response with optional data.
func (c *BaseController) SendSuccess(message string, data interface{}) {
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	c.Data["json"] = response{Success: true, Message: message, Data: data}
	c.ServeJSON()
	c.StopRun()
}

// SendCreated sends a 201 Created JSON response with optional data.
func (c *BaseController) SendCreated(message string, data interface{}) {
	c.Ctx.Output.SetStatus(201)
	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	c.Data["json"] = response{Success: true, Message: message, Data: data}
	c.ServeJSON()
	c.StopRun()
}

// SendError sends an error JSON response with the given HTTP status code.
func (c *BaseController) SendError(status int, message string) {
	c.Ctx.Output.SetStatus(status)
	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	c.Data["json"] = response{Success: false, Message: message}
	c.ServeJSON()
	c.StopRun()
}

// ParseBody unmarshals the request body JSON into the given target.
func (c *BaseController) ParseBody(target interface{}) error {
	return json.Unmarshal(c.Ctx.Input.RequestBody, target)
}

// GetAuthUserID extracts and validates the X-User-ID header.
// Returns the user ID and true on success, or 0 and false on failure.
func (c *BaseController) GetAuthUserID() (int, bool) {
	header := c.Ctx.Input.Header("X-User-ID")
	if header == "" {
		c.SendError(401, "Unauthorized")
		return 0, false
	}
	userID, err := strconv.Atoi(header)
	if err != nil || userID <= 0 {
		c.SendError(401, "Unauthorized")
		return 0, false
	}
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		c.SendError(401, "Unauthorized")
		return 0, false
	}
	return userID, true
}
