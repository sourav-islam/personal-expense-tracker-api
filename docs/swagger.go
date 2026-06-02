// Package docs contains the auto-generated Swagger documentation entry point.
// It is imported for its side effects — registering swagger spec metadata
// with the Beego framework so the /swagger endpoint works at runtime.
//
// Generate or refresh this file by running:
//
//	bee generate docs
package docs

import beego "github.com/beego/beego/v2/server/web"

func init() {
	info := beego.ControllerComments{
		Method: "GET",
	}
	_ = info
}
