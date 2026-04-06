package repository

import (
	"context"
	"nexsyn-backend/internal/database"
	"nexsyn-backend/internal/models"
)

type UserRepository struct{}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (id, fullname, email, password) VALUES ($1,$2,$3,$4) RETURNING id`

	return database.DB.QueryRow(context.Background(),
		query, user.ID, user.Fullname, user.Email, user.Password,
	).Scan(&user.ID)
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	rows, err := database.DB.Query(context.Background(), "SELECT id, fullname, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Fullname, &u.Email)
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) Update(id uint, user models.User) error {
	query := `UPDATE users SET fullname=$1, email=$2 WHERE id=$3`
	_, err := database.DB.Exec(context.Background(), query, user.Fullname, user.Email, id)
	return err
}

func (r *UserRepository) Delete(id uint) error {
	_, err := database.DB.Exec(context.Background(), "DELETE FROM users WHERE id=$1", id)
	return err
}