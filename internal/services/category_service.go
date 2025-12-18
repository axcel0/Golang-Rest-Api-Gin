package services

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"fmt"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(req *models.Category) (*models.Category, error) {
	if err := s.repo.Create(req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *CategoryService) GetCategory(id uint) (*models.Category, error) {
	return s.repo.GetByID(id)
}

func (s *CategoryService) ListCategories(page, limit int, search string) ([]models.Category, int64, error) {
	return s.repo.List(page, limit, search)
}

func (s *CategoryService) UpdateCategory(id uint, payload *models.Category) (*models.Category, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	existing.Name = payload.Name
	existing.Description = payload.Description

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	return s.repo.Delete(id)
}

// ValidateCategory ensures business rules
func (s *CategoryService) ValidateCategory(cat *models.Category) error {
	if cat.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
