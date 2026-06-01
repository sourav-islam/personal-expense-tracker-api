package controllers

import (
	"expense-tracker-api/models"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
)

// AuthMiddleware validates the X-User-ID header on every expense request.
// It returns the authenticated user ID, or sends an error response and returns 0.
func AuthMiddleware(c *BaseController) int {
	rawID := c.Ctx.Input.Header("X-User-ID")
	if rawID == "" {
		logs.Warn("AuthMiddleware: missing X-User-ID header")
		c.SendError(401, "Unauthorized")
		return 0
	}

	userID, err := strconv.Atoi(rawID)
	if err != nil || userID <= 0 {
		logs.Warn("AuthMiddleware: invalid X-User-ID value:", rawID)
		c.SendError(401, "Unauthorized")
		return 0
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		logs.Error("AuthMiddleware: error looking up user ID:", userID, err)
		c.SendError(500, "Internal server error")
		return 0
	}
	if user == nil {
		logs.Warn("AuthMiddleware: user ID not found:", userID)
		c.SendError(401, "Unauthorized")
		return 0
	}

	return userID
}
