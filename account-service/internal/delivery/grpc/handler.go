package grpc

import (
	"context"

	pb "BankingService/pb/account"
)

type AccountHandler struct {
	pb.UnimplementedAccountServiceServer
}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{}
}
 
func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	return &pb.AccountResponse{
		AccountId: "acc-777-demo",
		UserId:    req.UserId,
		Balance:   0.0,
		Status:    "ACTIVE",
	}, nil
}
