package repository


import (
    "context"
    "nexsyn-backend/internal/database"
    "nexsyn-backend/internal/models"
)

func CreateReview(review models.Review) error {
    query := `
    INSERT INTO reviews (name, email, service, rating, review)
    VALUES ($1, $2, $3, $4, $5)`

    _, err := database.DB.Exec(
        context.Background(),
        query,
        review.Name,
        review.Email,
        review.Service,
        review.Rating,
        review.Review,
    )
    return err
}

func GetReviews() ([]models.Review, error) {
    rows, err := database.DB.Query(
        context.Background(),
        "SELECT id, name, email, service, rating, review, created_at FROM reviews WHERE email = $1",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var reviews []models.Review
    for rows.Next() {
        var review models.Review

        err := rows.Scan(&review.ID, &review.Name, &review.Email, &review.Service, &review.Rating, &review.Review, &review.CreatedAt)
        if err != nil {
            return nil, err
        }
        reviews = append(reviews, review)
    }

    return reviews, nil

}