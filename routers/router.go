// Package routers registers all application routes.
package routers

import (
	"expense-tracker-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// Swagger UI static files
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// ── Actual working routes ──────────────────────────────────────────
	// Health
	beego.Router("/api/v1/health", &controllers.HealthController{}, "get:Get")

	// Auth
	beego.Router("/api/v1/auth/register", &controllers.AuthController{}, "post:Register")
	beego.Router("/api/v1/auth/login", &controllers.AuthController{}, "post:Login")

	// Expenses — summary BEFORE :id to avoid route conflict
	beego.Router("/api/v1/expenses/summary", &controllers.ExpenseController{}, "get:Summary")
	beego.Router("/api/v1/expenses", &controllers.ExpenseController{}, "post:Create;get:List")
	beego.Router("/api/v1/expenses/:id", &controllers.ExpenseController{}, "get:GetOne;put:Update;delete:Delete")

	// ── Namespace for swagger doc generation only ──────────────────────
	ns := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/auth",
			beego.NSInclude(
				&controllers.AuthController{},
			),
		),
		beego.NSNamespace("/expenses",
			beego.NSInclude(
				&controllers.ExpenseController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
