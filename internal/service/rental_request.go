package service

import (
	"context"
	"errors"
	"ticketprocessing/internal/messaging"
	"ticketprocessing/internal/models"
	"ticketprocessing/internal/repository"
	"time"
)

var (
	ErrRentalRequestNotFound = errors.New("rental request not found")
	ErrInvalidDateTime       = errors.New("invalid datetime format")
	ErrEquipmentNotFound     = errors.New("equipment not found")
	ErrInvalidDateRange      = errors.New("invalid date range")
)

type CreateRentalRequestRequest struct {
	EquipmentID uint      `json:"equipment_id"`
	FromDate    time.Time `json:"from_date"`
	ToDate      time.Time `json:"to_date"`
}

type RentalRequestService interface {
	GetRequestStatus(ctx context.Context, requestID uint) (*models.RequestStatusLog, error)
	GetRequestStatusAt(ctx context.Context, requestID uint, datetime time.Time) (*models.RequestStatusLog, error)
	CreateRentalRequest(ctx context.Context, userID uint, req CreateRentalRequestRequest) (*models.RentalRequest, error)
}

type rentalRequestService struct {
	rentalRequestRepo repository.RentalRequestRepository
	statusLogRepo     repository.RequestStatusLogRepository
	equipmentRepo     repository.EquipmentRepository
	publisher         messaging.RabbitMQPublisher
}

func NewRentalRequestService(
	rentalRequestRepo repository.RentalRequestRepository,
	statusLogRepo repository.RequestStatusLogRepository,
	equipmentRepo repository.EquipmentRepository,
	publisher messaging.RabbitMQPublisher,
) RentalRequestService {
	return &rentalRequestService{
		rentalRequestRepo: rentalRequestRepo,
		statusLogRepo:     statusLogRepo,
		equipmentRepo:     equipmentRepo,
		publisher:         publisher,
	}
}

func (s *rentalRequestService) GetRequestStatus(ctx context.Context, requestID uint) (*models.RequestStatusLog, error) {
	// Проверяем существование заявки
	request, err := s.rentalRequestRepo.GetRentalRequestByID(requestID)
	if err != nil {
		return nil, ErrRentalRequestNotFound
	}

	// Получаем последний статус заявки
	var statusLog models.RequestStatusLog
	if err := s.statusLogRepo.GetLatestStatusByRequestID(ctx, request.ID, &statusLog); err != nil {
		return nil, err
	}

	return &statusLog, nil
}

func (s *rentalRequestService) GetRequestStatusAt(ctx context.Context, requestID uint, datetime time.Time) (*models.RequestStatusLog, error) {
	request, err := s.rentalRequestRepo.GetRentalRequestByID(requestID)
	if err != nil {
		return nil, ErrRentalRequestNotFound
	}

	var statusLog models.RequestStatusLog
	if err := s.statusLogRepo.GetStatusAt(ctx, request.ID, datetime, &statusLog); err != nil {
		return nil, err
	}

	return &statusLog, nil
}

func (s *rentalRequestService) CreateRentalRequest(ctx context.Context, userID uint, req CreateRentalRequestRequest) (*models.RentalRequest, error) {
	// Проверяем существование оборудования
	equipment, err := s.equipmentRepo.GetEquipmentByID(req.EquipmentID)
	if err != nil {
		return nil, ErrEquipmentNotFound
	}

	// Проверяем даты
	if req.FromDate.After(req.ToDate) || req.FromDate.Before(time.Now()) {
		return nil, ErrInvalidDateRange
	}

	// Создаем заявку
	rentalRequest := &models.RentalRequest{
		UserID:      userID,
		EquipmentID: equipment.ID,
		FromDate:    req.FromDate,
		ToDate:      req.ToDate,
		Status:      "pending", // Начальный статус
	}

	if err := s.rentalRequestRepo.CreateRentalRequest(rentalRequest); err != nil {
		return nil, err
	}

	// Создаем начальный статус
	statusLog := &models.RequestStatusLog{
		RequestID: rentalRequest.ID,
		Status:    "pending",
		Timestamp: time.Now(),
		Comment:   "Request created",
	}

	if err := s.statusLogRepo.CreateRequestStatusLog(statusLog); err != nil {
		return nil, err
	}

	// Отправляем сообщение в RabbitMQ
	err = s.publisher.PublishRentalRequest(
		ctx,
		rentalRequest.ID,
		userID,
		equipment.ID,
		req.FromDate.Format(time.RFC3339),
		req.ToDate.Format(time.RFC3339),
	)
	if err != nil {
		// В случае ошибки отправки в RabbitMQ, мы все равно возвращаем созданную заявку
		// Возможно, стоит добавить механизм повторных попыток или компенсирующих транзакций
		return rentalRequest, nil
	}

	return rentalRequest, nil
}
