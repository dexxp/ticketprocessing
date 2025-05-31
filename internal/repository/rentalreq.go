package repository

import (
	"ticketprocessing/internal/models"

	"gorm.io/gorm"
)

type RentalRequestRepository interface {
	CreateRentalRequest(request *models.RentalRequest) error
	GetRentalRequestByID(id uint) (*models.RentalRequest, error)
	UpdateRentalRequest(request *models.RentalRequest) error
	DeleteRentalRequest(request *models.RentalRequest) error
}

type rentalRequestRepository struct {
	db *gorm.DB
}

func NewRentalRequestRepository(db *gorm.DB) RentalRequestRepository {
	return &rentalRequestRepository{db: db}
}

func (r *rentalRequestRepository) CreateRentalRequest(request *models.RentalRequest) error {
	return r.db.Create(request).Error
}

func (r *rentalRequestRepository) GetRentalRequestByID(id uint) (*models.RentalRequest, error) {
	var request models.RentalRequest
	if err := r.db.Where("id = ?", id).First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *rentalRequestRepository) UpdateRentalRequest(request *models.RentalRequest) error {
	return r.db.Save(request).Error
}

func (r *rentalRequestRepository) DeleteRentalRequest(request *models.RentalRequest) error {
	return r.db.Delete(request).Error
}
