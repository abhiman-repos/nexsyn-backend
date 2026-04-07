package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nexsyn-backend/internal/database"
	"nexsyn-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

func GoogleAuth(c *fiber.Ctx) error {
	type Request struct {
		Token string `json:"token"`
	}

	var body Request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	if body.Token == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Token required"})
	}

	payload, err := idtoken.Validate(
		context.Background(),
		body.Token,
		"207347033442-fhnk72ifibrtdovf67nrrtu1kjvb820p.apps.googleusercontent.com",
	)
	if err != nil {
		fmt.Println("TOKEN ERROR:", err)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid Google token"})
	}

	userID := payload.Subject

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		fmt.Println("EMAIL MISSING")
		return c.Status(400).JSON(fiber.Map{"error": "Email not found"})
	}

	fullname, _ := payload.Claims["name"].(string)
	if fullname == "" {
		fullname, _ = payload.Claims["given_name"].(string)
	}

	fmt.Println("USER:", userID, email, fullname)

	query := `
	INSERT INTO users (id, email, fullname, provider)
	VALUES ($1, $2, $3, 'google')
	ON CONFLICT (email)
	DO UPDATE SET
		fullname = EXCLUDED.fullname
	`

	_, err = database.DB.Exec(context.Background(),
		query,
		userID,
		email,
		fullname,
	)

	if err != nil {
		fmt.Println("DB ERROR:", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := utils.GenerateTokens(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token failed"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   true, // 🔥 must be true in production
		SameSite: "None",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Google login success",
	})
}

func AuthUser(c *fiber.Ctx) error {
	type Request struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Fullname string `json:"fullname"`
		Provider string `json:"provider"`
	}

	var body Request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	// ✅ Normalize input
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Fullname = strings.TrimSpace(body.Fullname)
	body.Provider = strings.TrimSpace(strings.ToLower(body.Provider))

	if body.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email required"})
	}

	if body.Provider == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Provider required"})
	}

	var userID string

	// ========================
	// 🔵 GOOGLE AUTH
	// ========================
	if body.Provider == "local" {

		// ========================
		// 🟢 LOCAL REGISTER
		// ========================
		if body.Password == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Password required"})
		}

		// 🔐 Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Hash failed"})
		}

		userID = uuid.New().String()

		query := `
		INSERT INTO users (id, email, password, fullname, provider)
		VALUES ($1, $2, $3, $4, 'local')
		`

		_, err = database.DB.Exec(context.Background(),
			query,
			userID,
			body.Email,
			string(hash),
			body.Fullname,
		)

		if err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				return c.Status(400).JSON(fiber.Map{"error": "Email already exists"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "User creation failed"})
		}

	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid provider"})
	}

	// ========================
	// 🔐 GENERATE TOKEN
	// ========================
	token, err := utils.GenerateTokens(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token failed"})
	}

	// 🍪 Secure cookie
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   true, // ⚠️ true in production (HTTPS)
		SameSite: "None",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Success",
		"userID":  userID,
	})
}

// 🔥 LOGIN USER
func LoginUser(c *fiber.Ctx) error {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body Request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// ✅ Normalize
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))

	if body.Email == "" || body.Password == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	var id string
	var hashedPassword string
	var provider string

	// 🔍 Fetch user
	query := `
	SELECT id, password, provider 
	FROM users 
	WHERE LOWER(email) = LOWER($1)
	`

	err := database.DB.QueryRow(context.Background(),
		query,
		body.Email,
	).Scan(&id, &hashedPassword, &provider)

	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// ❌ Prevent Google users from local login
	if provider == "google" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Use Google login",
		})
	}

	// 🔐 Compare password
	if err := utils.CheckPassword(hashedPassword, body.Password); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// 🎟️ Generate token
	token, err := utils.GenerateTokens(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// 🍪 Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   false, // true in production
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"user": fiber.Map{
			"id":    id,
			"email": body.Email,
		},
	})
}

// 🔥 GET CURRENT USER PROFILE (Only authenticated user's profile)
func GetProfile(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized - Invalid user context",
		})
	}

	// Query to get user profile with all fields
	query := `
	SELECT 
		id, 
		fullname, 
		email, 
		COALESCE(phone, '') as phone,
		COALESCE(gender, '') as gender,
		COALESCE(alternate_phone, '') as alternate_phone,
		created_at,
		updated_at
	FROM users 
	WHERE id = $1
	`

	var profile struct {
		ID             int       `json:"id"`
		FullName       string    `json:"fullname"`
		Email          string    `json:"email"`
		Phone          string    `json:"phone"`
		Gender         string    `json:"gender"`
		AlternatePhone string    `json:"alternatePhone"`
		CreatedAt      time.Time `json:"createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"`
	}

	err := database.DB.QueryRow(context.Background(), query, userID).Scan(
		&profile.ID,
		&profile.FullName,
		&profile.Email,
		&profile.Phone,
		&profile.Gender,
		&profile.AlternatePhone,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User profile not found",
		})
	}

	return c.JSON(profile)
}

// 🔥 UPDATE CURRENT USER PROFILE
func UpdateProfile(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Request structure
	type Request struct {
		FullName       string `json:"fullName"`
		Phone          string `json:"phone"`
		Gender         string `json:"gender"`
		AlternatePhone string `json:"alternatePhone"`
		Avatar         string `json:"avatar"`
	}

	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update query
	query := `
	UPDATE users 
	SET 
		fullname = COALESCE(NULLIF($1, ''), fullname),
		phone = COALESCE(NULLIF($2, ''), phone),
		gender = COALESCE(NULLIF($3, ''), gender),
		alternate_phone = COALESCE(NULLIF($4, ''), alternate_phone),
		avatar = COALESCE(NULLIF($5, ''), avatar),
		updated_at = $6
	WHERE id = $7
	RETURNING id, fullname, email, phone, avatar, gender, alternate_phone
	`

	var updatedProfile struct {
		ID             int    `json:"id"`
		Email          string `json:"email"`
		FullName       string `json:"fullName"`
		Phone          string `json:"phone"`
		Avatar         string `json:"avatar"`
		Gender         string `json:"gender"`
		AlternatePhone string `json:"alternatePhone"`
	}

	err := database.DB.QueryRow(
		context.Background(),
		query,
		body.FullName,
		body.Phone,
		body.Gender,
		body.AlternatePhone,
		time.Now(),
		userID,
	).Scan(
		&updatedProfile.ID,
		&updatedProfile.FullName,
		&updatedProfile.Email,
		&updatedProfile.Phone,
		&updatedProfile.Gender,
		&updatedProfile.AlternatePhone,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(updatedProfile)
}

// 🔥 GET ALL USERS (Admin only - keep for admin purposes)
func GetUsers(c *fiber.Ctx) error {
	rows, err := database.DB.Query(context.Background(),
		"SELECT id, fullname, email, phone, gender FROM users ORDER BY id",
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var users []fiber.Map

	for rows.Next() {
		var id int
		var fullName, email, phone, gender string

		rows.Scan(&id, &fullName, &email, &fullName, &phone, &gender)

		users = append(users, fiber.Map{
			"id":       id,
			"fullName": fullName,
			"email":    email,
			"phone":    phone,
			"gender":   gender,
		})
	}

	return c.JSON(users)
}

// 🔥 UPDATE USER (Admin only)
func UpdateUser(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	type Request struct {
		FullName string `json:"fullName"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Gender   string `json:"gender"`
	}

	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	_, err := database.DB.Exec(context.Background(),
		"UPDATE users SET email=$1, fullname=$2, phone=$3, gender=$4 WHERE id=$5",
		body.Email, body.FullName, body.Phone, body.Gender, id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User updated successfully"})
}

// 🔥 DELETE USER (Admin only)
func DeleteUser(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	_, err := database.DB.Exec(context.Background(),
		"DELETE FROM users WHERE id=$1",
		id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// 🔥 GET ME (Simple auth check)
func GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Authenticated",
		"userID":  userID,
	})
}

// 🔥 LOGOUT USER
func LogoutUser(c *fiber.Ctx) error {
	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
