// Package routers registers all application routes.
package routers

import (
	"expense-tracker-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

// Init registers all API routes for the application.
func init() {
	// Health
	beego.Router("/api/v1/health", &controllers.HealthController{}, "get:Get")

	// Auth
	beego.Router("/api/v1/auth/register", &controllers.AuthController{}, "post:Register")
	beego.Router("/api/v1/auth/login", &controllers.AuthController{}, "post:Login")
}
