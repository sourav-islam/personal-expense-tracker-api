package routers

import (
	"expense-tracker-api/controllers"

	"github.com/beego/beego/v2/server/web"
)

// Init initializes the routes.
func Init() {
	ns := web.NewNamespace("/api/v1",
		web.NSNamespace("/auth",
			web.NSRouter("/register", &controllers.AuthController{}, "post:Register"),
			web.NSRouter("/login", &controllers.AuthController{}, "post:Login"),
		),
		web.NSRouter("/health", &controllers.HealthController{}, "get:Get"),
	)
	web.AddNamespace(ns)
}
