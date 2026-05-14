package grpc

import (
	"context"
	"log"

	pb "BankingService/pb/transaction"

	"github.com/google/uuid"
)

type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

func (h *TransactionHandler) ProcessDeposit(ctx context.Context, req *pb.DepositRequest) (*pb.TransactionResponse, error) {
	_ = ctx

	transactionID := uuid.New().String()

	log.Printf("Deposit processed: %s | Account: %s | Amount: %f", transactionID, req.AccountId, req.Amount)

	return &pb.TransactionResponse{
		TransactionId: transactionID,
		Status:        "SUCCESS",
	}, nil
}
