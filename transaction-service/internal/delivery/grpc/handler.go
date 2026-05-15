package grpc

import (
	"context"
	"log"

	pb "BankingService/pb/transaction"
	"BankingService/transaction-service/internal/broker"
	"BankingService/transaction-service/internal/repository"

	"github.com/google/uuid"
)

type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	repo   *repository.TransactionRepository
	broker *broker.RabbitMQ
}

func NewTransactionHandler(repo *repository.TransactionRepository, broker *broker.RabbitMQ) *TransactionHandler {
	return &TransactionHandler{repo: repo, broker: broker}
}

func (h *TransactionHandler) ProcessDeposit(ctx context.Context, req *pb.DepositRequest) (*pb.TransactionResponse, error) {
	_ = ctx

	transactionID := uuid.New().String()

	err := h.repo.SaveTransaction(transactionID, req.AccountId, "DEPOSIT", req.Amount)
	if err != nil {
		log.Printf("Failed to save transaction: %v", err)
		return nil, err
	}

	err = h.broker.PublishDeposit(req.AccountId, req.Amount)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
	}

	log.Printf("Deposit processed: %s | Account: %s | Amount: %f", transactionID, req.AccountId, req.Amount)

	return &pb.TransactionResponse{
		TransactionId: transactionID,
		Status:        "SUCCESS",
	}, nil
}
