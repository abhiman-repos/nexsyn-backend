package handlers

import (
	"context"
	"nexsyn-backend/internal/database"
	"nexsyn-backend/internal/models"

	"github.com/gofiber/fiber/v2"
)

func CreateReview(c *fiber.Ctx) error {
	var review models.Review

	if err := c.BodyParser(&review); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// ✅ Check if email already exists
	var exists bool
	err := database.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM reviews WHERE email=$1)",
		review.Email,
	).Scan(&exists)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "database error",
		})
	}

	// 🔥 THIS WAS MISSING
	if exists {
		return c.Status(400).JSON(fiber.Map{
			"error": "You have already submitted a review",
		})
	}

	// ✅ Insert
	query := `
		INSERT INTO reviews (name, email, service, rating, review)
		VALUES ($1,$2,$3,$4,$5)
	`

	_, err = database.DB.Exec(
		context.Background(),
		query,
		review.Name,
		review.Email,
		review.Service,
		review.Rating,
		review.Review,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to create review",
		})
	}

	return c.JSON(fiber.Map{
		"message": "review added",
	})
}

func GetReviews(c *fiber.Ctx) error {

	rows, err := database.DB.Query(context.Background(),
		"SELECT id, name, email, service, rating, review FROM reviews ORDER BY id DESC",
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch reviews",
		})
	}

	defer rows.Close()

	var reviews []models.Review

	for rows.Next() {

		var r models.Review

		err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.Email,
			&r.Service,
			&r.Rating,
			&r.Review,
		)

		if err != nil {
			return err
		}

		reviews = append(reviews, r)
	}

	return c.JSON(reviews)
}