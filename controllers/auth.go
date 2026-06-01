package controllers

import (
	"encoding/json"
	"expense-tracker-api/models"
	"net/mail"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// AuthController handles authentication requests.
type AuthController struct {
	BaseController
}

// RegisterInput represents the input for user registration.
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput represents the input for user login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration.
// @Title Register
// @Summary Register a new user
// @Param body body controllers.RegisterInput true "Registration payload"
// @Success 201 {object} controllers.JSONResponse
// @Failure 400 {object} controllers.JSONResponse
// @Failure 409 {object} controllers.JSONResponse
// @Failure 500 {object} controllers.JSONResponse
// @router /api/v1/auth/register [post]
func (c *AuthController) Register() {
	var input RegisterInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Error("Failed to unmarshal registration input: %v", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if input.Name == "" {
		c.SendError(400, "Name is required")
		return
	}
	if input.Email == "" {
		c.SendError(400, "Email is required")
		return
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		c.SendError(400, "Invalid email format")
		return
	}
	if len(input.Password) < 6 {
		c.SendError(400, "Password must be at least 6 characters")
		return
	}

	existingUser, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Failed to check existing user: %v", err)
		c.SendError(500, "Failed to register user")
		return
	}
	if existingUser != nil {
		c.SendError(409, "Email already exists")
		return
	}

	nextID, err := models.GetNextUserID()
	if err != nil {
		logs.Error("Failed to get next user ID: %v", err)
		c.SendError(500, "Failed to register user")
		return
	}

	user := &models.User{
		ID:        nextID,
		Name:      input.Name,
		Email:     input.Email,
		Password:  input.Password,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := models.CreateUser(user); err != nil {
		logs.Error("Failed to create user: %v", err)
		c.SendError(500, "Failed to register user")
		return
	}

	c.SendSuccess(201, "User registered successfully", nil)
}

// Login handles user login.
// @Title Login
// @Summary User login
// @Param body body controllers.LoginInput true "Login payload"
// @Success 200 {object} controllers.JSONResponse
// @Failure 400 {object} controllers.JSONResponse
// @Failure 401 {object} controllers.JSONResponse
// @router /api/v1/auth/login [post]
func (c *AuthController) Login() {
	var input LoginInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Error("Failed to unmarshal login input: %v", err)
		c.SendError(400, "Invalid request body")
		return
	}

	if input.Email == "" {
		c.SendError(400, "Email is required")
		return
	}
	if input.Password == "" {
		c.SendError(400, "Password is required")
		return
	}

	user, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Failed to get user by email: %v", err)
		c.SendError(500, "Failed to login")
		return
	}

	if user == nil || user.Password != input.Password {
		c.SendError(401, "Invalid email or password")
		return
	}

	data := map[string]interface{}{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
	}

	c.SendSuccess(200, "Login successful", data)
}
