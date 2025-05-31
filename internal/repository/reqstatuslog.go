package repository

import (
	"context"
	"ticketprocessing/internal/models"
	"time"

	"gorm.io/gorm"
)

type RequestStatusLogRepository interface {
	CreateRequestStatusLog(log *models.RequestStatusLog) error
	GetRequestStatusLogByID(id uint) (*models.RequestStatusLog, error)
	UpdateRequestStatusLog(log *models.RequestStatusLog) error
	DeleteRequestStatusLog(log *models.RequestStatusLog) error
	GetLatestStatusByRequestID(ctx context.Context, requestID uint, log *models.RequestStatusLog) error
	GetStatusAt(ctx context.Context, requestID uint, datetime time.Time, log *models.RequestStatusLog) error
}

type requestStatusLogRepository struct {
	db *gorm.DB
}

func NewRequestStatusLogRepository(db *gorm.DB) RequestStatusLogRepository {
	return &requestStatusLogRepository{db: db}
}

func (r *requestStatusLogRepository) CreateRequestStatusLog(log *models.RequestStatusLog) error {
	return r.db.Create(log).Error
}

func (r *requestStatusLogRepository) GetRequestStatusLogByID(id uint) (*models.RequestStatusLog, error) {
	var log models.RequestStatusLog
	if err := r.db.Where("id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *requestStatusLogRepository) UpdateRequestStatusLog(log *models.RequestStatusLog) error {
	return r.db.Save(log).Error
}

func (r *requestStatusLogRepository) DeleteRequestStatusLog(log *models.RequestStatusLog) error {
	return r.db.Delete(log).Error
}

func (r *requestStatusLogRepository) GetLatestStatusByRequestID(ctx context.Context, requestID uint, log *models.RequestStatusLog) error {
	return r.db.Where("request_id = ?", requestID).
		Order("timestamp DESC").
		First(log).Error
}

func (r *requestStatusLogRepository) GetStatusAt(ctx context.Context, requestID uint, datetime time.Time, log *models.RequestStatusLog) error {
	return r.db.Where("request_id = ? AND timestamp <= ?", requestID, datetime).
		Order("timestamp DESC").
		First(log).Error
}
