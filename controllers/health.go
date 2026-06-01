package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

// HealthController handles the health check endpoint.
type HealthController struct {
	beego.Controller
}

// Get godoc
// @Title Health Check
// @Summary Returns server running status
// @Success 200 {object} map[string]interface{}
// @router /api/v1/health [get]
func (c *HealthController) Get() {
	logs.Info("Health check called")
	c.Data["json"] = map[string]interface{}{
		"success": true,
		"message": "Server is running",
	}
	c.ServeJSON()
}
