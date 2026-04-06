package server

import (
	"github.com/gofiber/fiber/v2"

	"nexsyn-backend/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "nexsyn-backend",
			AppName:      "nexsyn-backend",
		}),

		db: database.New(),
	}

	return server
}

