// Package routers registers all API routes under the /api/v1 namespace.
package routers

import (
	"expense-tracker-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	ns := beego.NewNamespace("/api/v1",

		// This will check the server health without requiring authentication
		beego.NSRouter("/health", &controllers.HealthController{}, "get:Get"),

		// Auth — no X-User-ID required
		beego.NSRouter("/auth/register", &controllers.AuthController{}, "post:Register"),
		beego.NSRouter("/auth/login", &controllers.AuthController{}, "post:Login"),

		// Expenses — all require X-User-ID header
		// /summary must be registered before /:id to avoid route conflict
		beego.NSRouter("/expenses/summary", &controllers.ExpenseController{}, "get:Summary"),
		beego.NSRouter("/expenses", &controllers.ExpenseController{}, "post:Create;get:List"),
		beego.NSRouter("/expenses/:id", &controllers.ExpenseController{}, "get:GetOne;put:Update;delete:Delete"),
	)

	beego.AddNamespace(ns)
}
