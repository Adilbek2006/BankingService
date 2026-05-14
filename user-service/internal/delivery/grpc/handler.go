package grpc

import (
	"context"
	"log"

	pb "BankingService/pb/user"
	"BankingService/user-service/internal/repository"

	"github.com/google/uuid"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	_ = ctx

	userID := uuid.New().String()

	err := h.repo.SaveUser(userID, req.Name, req.Email)
	if err != nil {
		log.Printf("Error saving user to the database: %v", err)
		return nil, err
	}

	log.Printf("A new user has registered: %s (%s)", req.Name, req.Email)

	return &pb.UserResponse{
		UserId: userID,
		Name:   req.Name,
		Email:  req.Email,
	}, nil
}
