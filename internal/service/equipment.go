package service

import (
	"context"
	"ticketprocessing/internal/models"
	"ticketprocessing/internal/repository"
)

type EquipmentService struct {
	repo repository.EquipmentRepository
}

func NewEquipment(repo repository.EquipmentRepository) *EquipmentService {
	return &EquipmentService{
		repo: repo,
	}
}

func (es *EquipmentService) Create(ctx context.Context, equipment *models.Equipment) error {
	return es.repo.CreateEquipment(equipment)
}

func (es *EquipmentService) GetByID(ctx context.Context, id uint) (*models.Equipment, error) {
	return es.repo.GetEquipmentByID(id)
}

func (es *EquipmentService) Update(ctx context.Context, equipment *models.Equipment) error {
	return es.repo.UpdateEquipment(equipment)
}

func (es *EquipmentService) Delete(ctx context.Context, equipment *models.Equipment) error {
	return es.repo.DeleteEquipment(equipment)
}
