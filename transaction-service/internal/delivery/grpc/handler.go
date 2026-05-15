package grpc

import (
	"context"
	"errors"
	"log"

	accountPb "BankingService/pb/account"
	pb "BankingService/pb/transaction"
	"BankingService/transaction-service/internal/broker"
	"BankingService/transaction-service/internal/repository"

	"github.com/google/uuid"
)

type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	repo          *repository.TransactionRepository
	broker        *broker.RabbitMQ
	accountClient accountPb.AccountServiceClient
}

func NewTransactionHandler(repo *repository.TransactionRepository, broker *broker.RabbitMQ, ac accountPb.AccountServiceClient) *TransactionHandler {
	return &TransactionHandler{
		repo:          repo,
		broker:        broker,
		accountClient: ac,
	}
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

func (h *TransactionHandler) InitiateTransfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransactionResponse, error) {
	accDetails, err := h.accountClient.GetAccountDetails(ctx, &accountPb.AccountIdRequest{
		AccountId: req.FromAccount,
	})
	if err != nil {
		return nil, err
	}

	if accDetails.Balance < req.Amount {
		return nil, errors.New("insufficient funds")
	}

	txID := uuid.New().String()

	err = h.repo.SaveTransfer(txID, req.FromAccount, req.ToAccount, req.Amount)
	if err != nil {
		return nil, err
	}

	err = h.broker.PublishTransfer(req.FromAccount, req.ToAccount, req.Amount)
	if err != nil {
		log.Printf("failed to publish transfer: %v", err)
	}

	return &pb.TransactionResponse{
		TransactionId: txID,
		Status:        "PENDING",
	}, nil
}
