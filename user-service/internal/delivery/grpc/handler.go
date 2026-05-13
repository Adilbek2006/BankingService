package grpc

import (
	"context"

	pb "BankingService/pb/user"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	return &pb.UserResponse{
		UserId: "12345-demo-id",
		Name:   req.Name,
		Email:  req.Email,
	}, nil
}
