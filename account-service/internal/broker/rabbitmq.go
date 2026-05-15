package broker

import (
	"BankingService/account-service/internal/repository"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	repo    *repository.AccountRepository
}

type DepositEvent struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

func NewRabbitMQ(url string, repo *repository.AccountRepository) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		"deposit_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{conn: conn, channel: ch, repo: repo}, nil
}

func (r *RabbitMQ) ConsumeDeposits() {
	msgs, err := r.channel.Consume(
		"deposit_queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Consumer error: %v", err)
	}

	go func() {
		for d := range msgs {
			var event DepositEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				continue
			}

			err := r.repo.UpdateBalance(event.AccountID, event.Amount)
			if err != nil {
				log.Printf("Balance update failed: %v", err)
			} else {
				log.Printf("Balance updated: Account %s | Added %f", event.AccountID, event.Amount)
			}
		}
	}()
}

func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
