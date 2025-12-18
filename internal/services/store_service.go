package services

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"fmt"
)

type StoreService struct {
	repo *repository.StoreRepository
}

func NewStoreService(repo *repository.StoreRepository) *StoreService {
	return &StoreService{repo: repo}
}

func (s *StoreService) CreateStore(req *models.Store) (*models.Store, error) {
	if err := s.repo.Create(req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *StoreService) GetStore(id uint) (*models.Store, error) {
	return s.repo.GetByID(id)
}

func (s *StoreService) ListStores(page, limit int, search string) ([]models.Store, int64, error) {
	return s.repo.List(page, limit, search)
}

func (s *StoreService) UpdateStore(id uint, payload *models.Store) (*models.Store, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	existing.Name = payload.Name
	existing.Address = payload.Address
	existing.PhoneNumber = payload.PhoneNumber
	existing.Email = payload.Email
	existing.IsActive = payload.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *StoreService) DeleteStore(id uint) error {
	return s.repo.Delete(id)
}

// ValidateStore ensures business rules
func (s *StoreService) ValidateStore(store *models.Store) error {
	if store.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
