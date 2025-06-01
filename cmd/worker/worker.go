package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"ticketprocessing/internal/config"
	"ticketprocessing/internal/db"
	"ticketprocessing/internal/models"
	"ticketprocessing/internal/repository"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rentalRequestMessage struct {
	RequestID   uint   `json:"request_id"`
	UserID      uint   `json:"user_id"`
	EquipmentID uint   `json:"equipment_id"`
	FromDate    string `json:"from_date"`
	ToDate      string `json:"to_date"`
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := db.InitPostgres(&cfg.Postgres)
	if err != nil {
		log.Error("failed to init postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}

	rentalRequestRepo := repository.NewRentalRequestRepository(db)
	statusLogRepo := repository.NewRequestStatusLogRepository(db)

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Error("failed to connect to RabbitMQ", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Error("failed to open channel", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		cfg.RabbitMQ.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("failed to declare queue", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		log.Error("failed to set QoS", slog.String("error", err.Error()))
		os.Exit(1)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("failed to register a consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	go func() {
		for msg := range msgs {
			processMessage(ctx, msg, rentalRequestRepo, statusLogRepo, log)
		}
		done()
	}()

	log.Info("Worker started, waiting for messages...")

	<-ctx.Done()
	log.Info("Shutting down worker...")

	if err := ch.Close(); err != nil {
		log.Error("failed to close channel", slog.String("error", err.Error()))
	}
	if err := conn.Close(); err != nil {
		log.Error("failed to close connection", slog.String("error", err.Error()))
	}

	log.Info("Worker stopped")
}

func processMessage(
	ctx context.Context,
	msg amqp.Delivery,
	rentalRequestRepo repository.RentalRequestRepository,
	statusLogRepo repository.RequestStatusLogRepository,
	log *slog.Logger,
) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic recovered while processing message",
				slog.Any("error", err),
				slog.String("message_id", msg.MessageId),
			)
			msg.Nack(false, false)
		}
	}()

	var requestMsg rentalRequestMessage
	if err := json.Unmarshal(msg.Body, &requestMsg); err != nil {
		log.Error("failed to unmarshal message",
			slog.String("error", err.Error()),
			slog.String("body", string(msg.Body)),
		)
		msg.Nack(false, false)
		return
	}

	log.Info("processing rental request",
		slog.Uint64("request_id", uint64(requestMsg.RequestID)),
		slog.Uint64("user_id", uint64(requestMsg.UserID)),
		slog.Uint64("equipment_id", uint64(requestMsg.EquipmentID)),
	)

	request, err := rentalRequestRepo.GetRentalRequestByID(requestMsg.RequestID)
	if err != nil {
		log.Error("failed to get rental request",
			slog.String("error", err.Error()),
			slog.Uint64("request_id", uint64(requestMsg.RequestID)),
		)
		msg.Nack(false, true)
		return
	}

	if request.Status != "pending" {
		log.Info("request already processed",
			slog.Uint64("request_id", uint64(requestMsg.RequestID)),
			slog.String("status", request.Status),
		)
		msg.Ack(false)
		return
	}

	// Здесь можно добавить дополнительную бизнес-логику обработки заявки
	// Например, проверка доступности оборудования, проверка кредитного рейтинга пользователя и т.д.

	// В данном примере просто меняем статус на "approved"
	request.Status = "approved"
	if err := rentalRequestRepo.UpdateRentalRequest(request); err != nil {
		log.Error("failed to update rental request",
			slog.String("error", err.Error()),
			slog.Uint64("request_id", uint64(requestMsg.RequestID)),
		)
		msg.Nack(false, true)
		return
	}

	statusLog := &models.RequestStatusLog{
		RequestID: request.ID,
		Status:    "approved",
		Timestamp: time.Now(),
		Comment:   "Request approved by worker",
	}

	if err := statusLogRepo.CreateRequestStatusLog(statusLog); err != nil {
		log.Error("failed to create status log",
			slog.String("error", err.Error()),
			slog.Uint64("request_id", uint64(requestMsg.RequestID)),
		)
	}

	log.Info("rental request processed successfully",
		slog.Uint64("request_id", uint64(requestMsg.RequestID)),
		slog.String("new_status", "approved"),
	)

	msg.Ack(false)
}
