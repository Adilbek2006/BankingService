package grpc

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	pb "BankingService/pb/user"
	"BankingService/user-service/internal/email"
	"BankingService/user-service/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	repo   *repository.UserRepository
	sender email.Sender
}

func NewUserHandler(repo *repository.UserRepository, sender email.Sender) *UserHandler {
	if sender == nil {
		sender = email.NewNoopSender()
	}
	return &UserHandler{repo: repo, sender: sender}
}

const (
	tokenVerifyEmail   = "VERIFY_EMAIL"
	tokenResetPassword = "RESET_PASSWORD"
)

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	_ = ctx

	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	userID := uuid.New().String()
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	err = h.repo.SaveUser(userID, req.Name, req.Email, passwordHash)
	if err != nil {
		log.Printf("Error saving user to the database: %v", err)
		return nil, err
	}

	verifyToken, err := generateToken()
	if err != nil {
		return nil, err
	}
	if err := h.repo.CreateToken(verifyToken, userID, tokenVerifyEmail, time.Now().Add(24*time.Hour)); err != nil {
		return nil, err
	}

	subject := "Verify your email"
	body := fmt.Sprintf("Hello,\r\n\r\nYour verification token:\r\n%s\r\n\r\nIf you did not request this, ignore this email.\r\n", verifyToken)
	if err := h.sender.Send(req.Email, subject, body); err != nil {
		return nil, err
	}

	log.Printf("A new user has registered: %s (%s)", req.Name, req.Email)

	return &pb.UserResponse{
		UserId: userID,
		Name:   req.Name,
		Email:  req.Email,
	}, nil
}

func (h *UserHandler) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	_ = ctx

	if err := h.repo.UpdateUserName(req.UserId, req.Name); err != nil {
		return nil, err
	}

	user, err := h.repo.GetUserByID(req.UserId)
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
	_ = ctx

	user, err := h.repo.GetUserByID(req.UserId)
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
	_ = ctx

	if err := h.repo.DeleteUser(req.UserId); err != nil {
		return &pb.StatusResponse{Success: false, Message: "delete_failed"}, err
	}

	return &pb.StatusResponse{Success: true, Message: "deleted"}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.StatusResponse, error) {
	_ = ctx

	user, err := h.repo.GetUserByID(req.UserId)
	if err != nil {
		return nil, err
	}
	if !user.PasswordHash.Valid {
		return nil, errors.New("password not set")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.OldPassword)); err != nil {
		return &pb.StatusResponse{Success: false, Message: "invalid_password"}, nil
	}

	newHash, err := hashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	if err := h.repo.UpdatePasswordHash(req.UserId, newHash); err != nil {
		return &pb.StatusResponse{Success: false, Message: "update_failed"}, err
	}

	return &pb.StatusResponse{Success: true, Message: "password_changed"}, nil
}

func (h *UserHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.StatusResponse, error) {
	_ = ctx

	okToken, err := h.repo.GetValidToken(req.Token, tokenVerifyEmail)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: "invalid_token"}, nil
	}

	if err := h.repo.SetEmailVerified(okToken.UserID, true); err != nil {
		return &pb.StatusResponse{Success: false, Message: "verify_failed"}, err
	}

	_ = h.repo.MarkTokenUsed(okToken.Token)

	return &pb.StatusResponse{Success: true, Message: "email_verified"}, nil
}

func (h *UserHandler) SendPasswordReset(ctx context.Context, req *pb.EmailRequest) (*pb.StatusResponse, error) {
	_ = ctx

	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pb.StatusResponse{Success: true, Message: "reset_sent"}, nil
		}
		return nil, err
	}

	ok, err := generateToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(30 * time.Minute)
	if err := h.repo.CreateToken(ok, user.UserID, tokenResetPassword, expiresAt); err != nil {
		return nil, err
	}

	subject := "Password reset"
	body := fmt.Sprintf("Hello,\r\n\r\nYour password reset token:\r\n%s\r\n\r\nIf you did not request this, ignore this email.\r\n", ok)
	if err := h.sender.Send(req.Email, subject, body); err != nil {
		return nil, err
	}

	return &pb.StatusResponse{Success: true, Message: "reset_sent"}, nil
}

func (h *UserHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.StatusResponse, error) {
	_ = ctx

	token, err := h.repo.GetValidToken(req.Token, tokenResetPassword)
	if err != nil {
		return &pb.StatusResponse{Success: false, Message: "invalid_token"}, nil
	}

	newHash, err := hashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	if err := h.repo.UpdatePasswordHash(token.UserID, newHash); err != nil {
		return &pb.StatusResponse{Success: false, Message: "update_failed"}, err
	}

	_ = h.repo.MarkTokenUsed(token.Token)

	return &pb.StatusResponse{Success: true, Message: "password_reset"}, nil
}

func (h *UserHandler) UpdateKYCStatus(ctx context.Context, req *pb.UpdateKYCRequest) (*pb.StatusResponse, error) {
	_ = ctx

	if err := h.repo.UpdateKYCStatus(req.UserId, req.Status); err != nil {
		return &pb.StatusResponse{Success: false, Message: "kyc_update_failed"}, err
	}

	return &pb.StatusResponse{Success: true, Message: "kyc_updated"}, nil
}

func (h *UserHandler) GetKYCStatus(ctx context.Context, req *pb.UserIdRequest) (*pb.KYCResponse, error) {
	_ = ctx

	user, err := h.repo.GetUserByID(req.UserId)
	if err != nil {
		return nil, err
	}

	status := ""
	if user.KYCStatus.Valid {
		status = user.KYCStatus.String
	}

	return &pb.KYCResponse{Status: status}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, _ *pb.EmptyRequest) (*pb.UserListResponse, error) {
	_ = ctx

	users, err := h.repo.ListUsers()
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
	_ = ctx

	if err := h.repo.SetSuspended(req.UserId, true); err != nil {
		return &pb.StatusResponse{Success: false, Message: "suspend_failed"}, err
	}

	return &pb.StatusResponse{Success: true, Message: "suspended"}, nil
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
