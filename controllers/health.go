package controllers

// HealthController handles health check requests.
type HealthController struct {
	BaseController
}

// Get handles the health check request.
// @Title Get
// @Summary Health check
// @Success 200 {object} controllers.JSONResponse
// @router /api/v1/health [get]
func (c *HealthController) Get() {
	c.SendSuccess(200, "Server is running", nil)
}
