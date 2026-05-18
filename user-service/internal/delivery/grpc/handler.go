package grpc

import (
	"context"

	pb "BankingService/pb/user"
	"BankingService/user-service/internal/usecase"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	userID, err := h.uc.CreateUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		UserId: userID,
		Name:   req.Name,
		Email:  req.Email,
	}, nil
}

func (h *UserHandler) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	user, err := h.uc.UpdateUserProfile(ctx, req.UserId, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		UserId: user.UserID,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

func (h *UserHandler) GetUserById(ctx context.Context, req *pb.UserIdRequest) (*pb.UserResponse, error) {
	user, err := h.uc.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		UserId: user.UserID,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.UserIdRequest) (*pb.StatusResponse, error) {
	err := h.uc.DeleteUser(ctx, req.UserId)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: "delete_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "deleted"}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.StatusResponse, error) {
	success, message, err := h.uc.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: message}, err
	}
	return &pb.StatusResponse{Success: success, Message: message}, nil
}

func (h *UserHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.StatusResponse, error) {
	success, message, err := h.uc.VerifyEmail(ctx, req.Token)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: message}, err
	}
	return &pb.StatusResponse{Success: success, Message: message}, nil
}

func (h *UserHandler) SendPasswordReset(ctx context.Context, req *pb.EmailRequest) (*pb.StatusResponse, error) {
	success, message, err := h.uc.SendPasswordReset(ctx, req.Email)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: message}, err
	}
	return &pb.StatusResponse{Success: success, Message: message}, nil
}

func (h *UserHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.StatusResponse, error) {
	success, message, err := h.uc.ResetPassword(ctx, req.Token, req.NewPassword)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: message}, err
	}
	return &pb.StatusResponse{Success: success, Message: message}, nil
}

func (h *UserHandler) UpdateKYCStatus(ctx context.Context, req *pb.UpdateKYCRequest) (*pb.StatusResponse, error) {
	success, message, err := h.uc.UpdateKYCStatus(ctx, req.UserId, req.Status)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: message}, err
	}
	return &pb.StatusResponse{Success: success, Message: message}, nil
}

func (h *UserHandler) GetKYCStatus(ctx context.Context, req *pb.UserIdRequest) (*pb.KYCResponse, error) {
	status, err := h.uc.GetKYCStatus(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.KYCResponse{Status: status}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, _ *pb.EmptyRequest) (*pb.UserListResponse, error) {
	users, err := h.uc.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	resp := &pb.UserListResponse{}
	for _, u := range users {
		resp.Users = append(resp.Users, &pb.UserResponse{
			UserId: u.UserID,
			Name:   u.Name,
			Email:  u.Email,
		})
	}

	return resp, nil
}

func (h *UserHandler) SuspendUser(ctx context.Context, req *pb.UserIdRequest) (*pb.StatusResponse, error) {
	success, message, err := h.uc.SuspendUser(ctx, req.UserId)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: message}, err
	}
	return &pb.StatusResponse{Success: success, Message: message}, nil
}
