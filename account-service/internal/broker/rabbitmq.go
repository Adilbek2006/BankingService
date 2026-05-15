package broker

import (
	"context"
	"encoding/json"
	"log"

	"BankingService/account-service/internal/repository"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	repo    *repository.AccountRepository
	cache   *repository.RedisCache
}

type DepositEvent struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type TransferEvent struct {
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      float64 `json:"amount"`
}

func NewRabbitMQ(url string, repo *repository.AccountRepository, cache *repository.RedisCache) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare("deposit_queue", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare("transfer_queue", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{conn: conn, channel: ch, repo: repo, cache: cache}, nil
}

func (r *RabbitMQ) ConsumeDeposits() {
	msgs, err := r.channel.Consume("deposit_queue", "", true, false, false, false, nil)
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
			if err == nil {
				_ = r.cache.DeleteFromCache(context.Background(), event.AccountID)
				log.Printf("Balance updated: Account %s | Added %f", event.AccountID, event.Amount)
			}
		}
	}()
}

func (r *RabbitMQ) ConsumeTransfers() {
	msgs, err := r.channel.Consume("transfer_queue", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Consumer error: %v", err)
	}

	go func() {
		for d := range msgs {
			var event TransferEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				continue
			}

			_ = r.repo.UpdateBalance(event.FromAccount, -event.Amount)
			_ = r.repo.UpdateBalance(event.ToAccount, event.Amount)

			_ = r.cache.DeleteFromCache(context.Background(), event.FromAccount)
			_ = r.cache.DeleteFromCache(context.Background(), event.ToAccount)

			log.Printf("Transfer processed: %f from %s to %s", event.Amount, event.FromAccount, event.ToAccount)
		}
	}()
}

func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
