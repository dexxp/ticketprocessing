package repository

import (
	"ticketprocessing/internal/models"

	"gorm.io/gorm"
)

type EquipmentRepository interface {
	CreateEquipment(equipment *models.Equipment) error
	GetEquipmentByID(id uint) (*models.Equipment, error)
	UpdateEquipment(equipment *models.Equipment) error
	DeleteEquipment(equipment *models.Equipment) error
}

type equipmentRepository struct {
	db *gorm.DB
}

func NewEquipmentRepository(db *gorm.DB) EquipmentRepository {
	return &equipmentRepository{db: db}
}

func (r *equipmentRepository) CreateEquipment(equipment *models.Equipment) error {
	return r.db.Create(equipment).Error
}

func (r *equipmentRepository) GetEquipmentByID(id uint) (*models.Equipment, error) {
	var equipment models.Equipment
	if err := r.db.Where("id = ?", id).First(&equipment).Error; err != nil {
		return nil, err
	}
	return &equipment, nil
}

func (r *equipmentRepository) UpdateEquipment(equipment *models.Equipment) error {
	return r.db.Save(equipment).Error
}

func (r *equipmentRepository) DeleteEquipment(equipment *models.Equipment) error {
	return r.db.Delete(equipment).Error
}
