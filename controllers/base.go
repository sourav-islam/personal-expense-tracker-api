package controllers

import (
	"github.com/beego/beego/v2/server/web"
)

// BaseController is the base controller for all other controllers.
type BaseController struct {
	web.Controller
}

// JSONResponse is the standard JSON response format.
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendSuccess sends a standard success JSON response.
func (c *BaseController) SendSuccess(status int, message string, data interface{}) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.ServeJSON()
}

// SendError sends a standard error JSON response.
func (c *BaseController) SendError(status int, message string) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = JSONResponse{
		Success: false,
		Message: message,
	}
	c.ServeJSON()
}
