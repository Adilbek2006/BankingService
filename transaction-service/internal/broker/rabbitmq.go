package broker

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
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

func NewRabbitMQ(url string) (*RabbitMQ, error) {
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

	log.Println("RabbitMQ Publisher connected")
	return &RabbitMQ{conn: conn, channel: ch}, nil
}

func (r *RabbitMQ) PublishDeposit(accountID string, amount float64) error {
	event := DepositEvent{
		AccountID: accountID,
		Amount:    amount,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.channel.Publish("", "deposit_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (r *RabbitMQ) PublishTransfer(fromAccount, toAccount string, amount float64) error {
	event := TransferEvent{
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.channel.Publish("", "transfer_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.conn.Close()
}
