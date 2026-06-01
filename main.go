package main

import (
	"expense-tracker-api/routers"
	"os"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

func main() {
	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		logs.Error("Failed to create data directory: %v", err)
	}

	routers.Init()
	web.Run()
}
