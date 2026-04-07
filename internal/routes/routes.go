package routes

import (
	"nexsyn-backend/internal/handlers"
	"nexsyn-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func RegisterFiberRoutes(app *fiber.App) {

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// ✅ OPTIONS FIX
	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	api := app.Group("/api")

	// REVIEWS
	reviews := api.Group("/reviews")
	reviews.Get("/", handlers.GetReviews)
	reviews.Post("/", handlers.CreateReview)

	// USERS
	users := api.Group("/users")
	users.Post("/auth", handlers.AuthUser)
	users.Post("/google", handlers.GoogleAuth)
	users.Get("/me", middleware.AuthRequired(), handlers.GetMe)
	users.Get("/profile", handlers.GetProfile)
	
	users.Post("/login", handlers.LoginUser)
	users.Post("/logout", handlers.LogoutUser)
	users.Put("/profile/:id", middleware.AuthRequired(), handlers.UpdateUser)
	users.Delete("/profile/:id", middleware.AuthRequired(), handlers.DeleteUser)
}