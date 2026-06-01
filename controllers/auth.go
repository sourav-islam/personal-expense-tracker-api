package controllers

import (
	"encoding/json"
	"expense-tracker-api/models"

	"github.com/beego/beego/v2/core/logs"
)

// AuthController handles user registration and login.
type AuthController struct {
	BaseController
}

// registerInput defines the expected JSON body for registration.
type registerInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginInput defines the expected JSON body for login.
type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginData is the data payload returned on successful login.
type loginData struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// Register godoc
// @Title Register
// @Summary Register a new user account
// @Description Creates a new user with name, email, and password
// @Param body body controllers.registerInput true "Registration payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @router /api/v1/auth/register [post]
func (c *AuthController) Register() {
	logs.Info("Register endpoint called")

	var input registerInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Register: failed to parse request body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	// Validate name
	if input.Name == "" {
		c.SendError(400, "Name is required")
		return
	}

	// Validate email
	if input.Email == "" {
		c.SendError(400, "Email is required")
		return
	}
	if !models.ValidateEmail(input.Email) {
		c.SendError(400, "Invalid email format")
		return
	}

	// Validate password
	if input.Password == "" {
		c.SendError(400, "Password is required")
		return
	}
	if len(input.Password) < 6 {
		c.SendError(400, "Password must be at least 6 characters")
		return
	}

	// Check duplicate email
	existing, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Register: error checking existing email:", err)
		c.SendError(500, "Failed to register user")
		return
	}
	if existing != nil {
		logs.Warn("Register: duplicate email attempt:", input.Email)
		c.SendError(409, "Email already exists")
		return
	}

	// Create the user
	newUser := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}
	if err := models.CreateUser(newUser); err != nil {
		logs.Error("Register: failed to create user:", err)
		c.SendError(500, "Failed to register user")
		return
	}

	logs.Info("Register: new user created, email:", input.Email)
	c.SendSuccess(201, "User registered successfully", nil)
}

// Login godoc
// @Title Login
// @Summary Log in with email and password
// @Description Authenticates a user and returns their profile data
// @Param body body controllers.loginInput true "Login payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @router /api/v1/auth/login [post]
func (c *AuthController) Login() {
	logs.Info("Login endpoint called")

	var input loginInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Login: failed to parse request body:", err)
		c.SendError(400, "Invalid request body")
		return
	}

	// Validate fields
	if input.Email == "" {
		c.SendError(400, "Email is required")
		return
	}
	if input.Password == "" {
		c.SendError(400, "Password is required")
		return
	}

	// Find user by email
	user, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Login: error looking up user:", err)
		c.SendError(500, "Internal server error")
		return
	}
	if user == nil || user.Password != input.Password {
		logs.Warn("Login: invalid credentials for email:", input.Email)
		c.SendError(401, "Invalid email or password")
		return
	}

	logs.Info("Login: successful for user ID:", user.ID)
	c.SendSuccess(200, "Login successful", loginData{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
	})
}
