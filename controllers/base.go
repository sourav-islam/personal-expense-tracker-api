// Package controllers handles all incoming HTTP requests.
package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
)

// BaseController embeds beego.Controller and provides
// shared response helpers for all other controllers.
type BaseController struct {
	beego.Controller
}

// successResponse is the standard JSON structure for successful responses.
type successResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// errorResponse is the standard JSON structure for error responses.
type errorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendSuccess writes a successful JSON response with the given
// HTTP status code, message, and optional data payload.
func (c *BaseController) SendSuccess(status int, message string, data interface{}) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = successResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.ServeJSON()
}

// SendError writes a JSON error response with the given
// HTTP status code and message.
func (c *BaseController) SendError(status int, message string) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = errorResponse{
		Success: false,
		Message: message,
	}
	c.ServeJSON()
}
