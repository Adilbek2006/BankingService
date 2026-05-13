package grpc

import (
	"context"
	"log"

	"BankingService/account-service/internal/repository"
	pb "BankingService/pb/account"
	"github.com/google/uuid"
)

type AccountHandler struct {
	pb.UnimplementedAccountServiceServer
	repo *repository.AccountRepository
}

func NewAccountHandler(repo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	accountID := uuid.New().String()

	err := h.repo.SaveAccount(accountID, req.UserId, req.Currency)
	if err != nil {
		log.Printf(" Saving an error to the database: %v", err)
		return nil, err
	}

	log.Printf("Created new count: %s for user: %s", accountID, req.UserId)

	return &pb.AccountResponse{
		AccountId: accountID,
		UserId:    req.UserId,
		Balance:   0.0,
		Status:    "ACTIVE",
	}, nil
}
