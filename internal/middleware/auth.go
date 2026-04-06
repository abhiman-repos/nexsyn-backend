// internal/middleware/auth.go
package middleware

import (
	"strings"

	"nexsyn-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check cookie first
		token := c.Cookies("token")
		
		// If no token in cookie, check Authorization header
		if token == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}
		
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - No token provided",
			})
		}
		
		// Verify token and extract user ID
		userID, err := utils.VerifyToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - Invalid token",
			})
		}
		
		// Store user ID in context for later use
		c.Locals("userID", userID)
		
		return c.Next()
	}
}