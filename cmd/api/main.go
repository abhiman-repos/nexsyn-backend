package main

import (
	"log"

	"nexsyn-backend/internal/database"
	"nexsyn-backend/internal/server"
	"nexsyn-backend/internal/routes"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	if err := database.ConnectDB(); err != nil {
	
		log.Fatal(err)
	}

	if err := database.CreateTables(); err != nil {
		log.Fatal("Error creating tables:", err)
	} else {
		log.Println("✅ Tables created successfully")
	}

	s := server.New()

	s.App.Use(logger.New())

	// ✅ CALL ROUTES HERE (IMPORTANT)
	routes.RegisterFiberRoutes(s.App)

	log.Println("🚀 Server running on http://localhost:8000")

	log.Fatal(s.App.Listen(":8000"))
}

