package controllers

// HealthController handles the server health check endpoint.
type HealthController struct {
	BaseController
}

// Get returns the current server status.
// Title HealthCheck
// Summary Server health check
// Description Returns a simple status message confirming the server is running
// Success 200 {object} map[string]interface{} "Server is running"
// router /health [get]
func (c *HealthController) Get() {
	c.SendSuccess("Server is running", nil)
}
