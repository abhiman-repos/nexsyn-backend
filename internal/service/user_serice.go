package service

import (
	"nexsyn-backend/internal/models"
	"nexsyn-backend/internal/repository"
	"nexsyn-backend/internal/utils"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) Register(user *models.User) error {
	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = hash
	return s.repo.Create(user)
}

func (s *UserService) GetUsers() ([]models.User, error) {
	return s.repo.GetAll()
}

func (s *UserService) UpdateUser(id uint, user models.User) error {
	return s.repo.Update(id, user)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}