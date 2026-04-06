package main

import (
	"log"

	"nexsyn-backend/internal/database"
	"nexsyn-backend/internal/server"
	"nexsyn-backend/internal/routes"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	s := server.New()

	s.App.Use(logger.New())

	// ✅ CALL ROUTES HERE (IMPORTANT)
	routes.RegisterFiberRoutes(s.App)

	log.Println("🚀 Server running on http://localhost:8000")

	log.Fatal(s.App.Listen(":8000"))
}

