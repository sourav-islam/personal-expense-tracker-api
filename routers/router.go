// Package routers registers all application routes.
package routers

import (
	"expense-tracker-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

// init registers all API routes for the application.
func init() {
	// Enable Swagger UI in dev mode
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	// Health
	beego.Router("/api/v1/health", &controllers.HealthController{}, "get:Get")

	// Auth
	beego.Router("/api/v1/auth/register", &controllers.AuthController{}, "post:Register")
	beego.Router("/api/v1/auth/login", &controllers.AuthController{}, "post:Login")

	// Expenses — summary MUST come before :id to avoid route conflict
	beego.Router("/api/v1/expenses/summary", &controllers.ExpenseController{}, "get:Summary")
	beego.Router("/api/v1/expenses", &controllers.ExpenseController{}, "post:Create;get:List")
	beego.Router("/api/v1/expenses/:id", &controllers.ExpenseController{}, "get:GetOne;put:Update;delete:Delete")
}
