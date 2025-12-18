package repository

import (
	"Go-Lang-project-01/internal/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type StoreRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) *StoreRepository {
	return &StoreRepository{db: db}
}

func (r *StoreRepository) Create(store *models.Store) error {
	if err := r.db.Create(store).Error; err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	return nil
}

func (r *StoreRepository) GetByID(id uint) (*models.Store, error) {
	var store models.Store
	if err := r.db.First(&store, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("store with ID %d not found", id)
		}
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepository) List(page, limit int, search string) ([]models.Store, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var stores []models.Store
	var total int64

	query := r.db.Model(&models.Store{})

	if search != "" {
		pattern := fmt.Sprintf("%%%s%%", search)
		query = query.Where("name LIKE ? OR address LIKE ?", pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count stores: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&stores).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list stores: %w", err)
	}

	return stores, total, nil
}

func (r *StoreRepository) Update(store *models.Store) error {
	if err := r.db.Save(store).Error; err != nil {
		return fmt.Errorf("failed to update store: %w", err)
	}
	return nil
}

func (r *StoreRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Store{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("store with ID %d not found", id)
	}
	return nil
}
