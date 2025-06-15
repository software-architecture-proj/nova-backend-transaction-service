package notification

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

const (
	amqpURL     = "amqp://notifier:S3cUr0!Pass@localhost:5672/notifications"
	exchange    = "notifications"
	routingKey  = "email"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type TransactionNotification struct {
	Type          string  `json:"type"`
	Email         string  `json:"email"`
	TransactionID string  `json:"transactionId"`
	Amount        float64 `json:"amount"`
}

func NewProducer() (*Producer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// Declare the exchange
	err = ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %v", err)
	}

	// Declare the queue
	q, err := ch.QueueDeclare(
		"",    // name (empty means random name)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,    // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %v", err)
	}

	err = ch.Confirm(false)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to put channel in confirm mode: %v", err)
	}

	return &Producer{
		conn:    conn,
		channel: ch,
	}, nil
}

func (p *Producer) SendTransactionNotification(email, transactionID string, amount float64) error {
	notification := TransactionNotification{
		Type:          "transaction",
		Email:         email,
		TransactionID: transactionID,
		Amount:        amount,
	}

	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	err = p.channel.Publish(
		exchange,    // exchange
		routingKey,  // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    fmt.Sprintf("%d", time.Now().UnixNano()),
			Timestamp:    time.Now(),
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
} 
