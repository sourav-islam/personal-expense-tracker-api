// Package docs registers Swagger API metadata.
//
// @title          Personal Expense Tracker API
// @version        1.0
// @description    RESTful API for tracking personal expenses using Go and Beego v2.
// @host           localhost:8080
// @BasePath       /
package docs

import beego "github.com/beego/beego/v2/server/web"

func init() {
	info := beego.ControllerComments{
		Method: "GET",
	}
	_ = info
}
