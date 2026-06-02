package main

import (
	_ "expense-tracker-api/docs" // registers swagger docs
	_ "expense-tracker-api/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	beego.Run()
}
