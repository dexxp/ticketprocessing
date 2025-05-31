package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"ticketprocessing/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher interface {
	PublishRentalRequest(ctx context.Context, requestID uint, userID uint, equipmentID uint, fromDate, toDate string) error
	Close() error
}

type rabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewRabbitMQPublisher(cfg *config.RabbitMQConfig) (RabbitMQPublisher, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Объявляем очередь
	_, err = ch.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &rabbitMQPublisher{
		conn:    conn,
		channel: ch,
		queue:   cfg.Queue,
	}, nil
}

type rentalRequestMessage struct {
	RequestID   uint   `json:"request_id"`
	UserID      uint   `json:"user_id"`
	EquipmentID uint   `json:"equipment_id"`
	FromDate    string `json:"from_date"`
	ToDate      string `json:"to_date"`
}

func (p *rabbitMQPublisher) PublishRentalRequest(ctx context.Context, requestID uint, userID uint, equipmentID uint, fromDate, toDate string) error {
	msg := rentalRequestMessage{
		RequestID:   requestID,
		UserID:      userID,
		EquipmentID: equipmentID,
		FromDate:    fromDate,
		ToDate:      toDate,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.channel.PublishWithContext(ctx,
		"",      // exchange
		p.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (p *rabbitMQPublisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}
