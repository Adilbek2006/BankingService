package usecase

import (
	"context"

	userPb "BankingService/pb/user"
)

type UserUsecase struct {
	client userPb.UserServiceClient
}

func NewUserUsecase(client userPb.UserServiceClient) *UserUsecase {
	return &UserUsecase{client: client}
}

func (u *UserUsecase) CreateUser(ctx context.Context, name, email, password string) (*userPb.UserResponse, error) {
	return u.client.CreateUser(ctx, &userPb.CreateUserRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})
}

func (u *UserUsecase) UpdateUserProfile(ctx context.Context, userID, name string) (*userPb.UserResponse, error) {
	return u.client.UpdateUserProfile(ctx, &userPb.UpdateUserRequest{UserId: userID, Name: name})
}

func (u *UserUsecase) GetUserByID(ctx context.Context, userID string) (*userPb.UserResponse, error) {
	return u.client.GetUserById(ctx, &userPb.UserIdRequest{UserId: userID})
}

func (u *UserUsecase) DeleteUser(ctx context.Context, userID string) (*userPb.StatusResponse, error) {
	return u.client.DeleteUser(ctx, &userPb.UserIdRequest{UserId: userID})
}

func (u *UserUsecase) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) (*userPb.StatusResponse, error) {
	return u.client.ChangePassword(ctx, &userPb.ChangePasswordRequest{
		UserId:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
}

func (u *UserUsecase) VerifyEmail(ctx context.Context, userID, token string) (*userPb.StatusResponse, error) {
	return u.client.VerifyEmail(ctx, &userPb.VerifyEmailRequest{UserId: userID, Token: token})
}

func (u *UserUsecase) SendPasswordReset(ctx context.Context, email string) (*userPb.StatusResponse, error) {
	return u.client.SendPasswordReset(ctx, &userPb.EmailRequest{Email: email})
}

func (u *UserUsecase) ResetPassword(ctx context.Context, token, newPassword string) (*userPb.StatusResponse, error) {
	return u.client.ResetPassword(ctx, &userPb.ResetPasswordRequest{Token: token, NewPassword: newPassword})
}

func (u *UserUsecase) UpdateKYCStatus(ctx context.Context, userID, status string) (*userPb.StatusResponse, error) {
	return u.client.UpdateKYCStatus(ctx, &userPb.UpdateKYCRequest{UserId: userID, Status: status})
}

func (u *UserUsecase) GetKYCStatus(ctx context.Context, userID string) (*userPb.KYCResponse, error) {
	return u.client.GetKYCStatus(ctx, &userPb.UserIdRequest{UserId: userID})
}

func (u *UserUsecase) ListUsers(ctx context.Context) (*userPb.UserListResponse, error) {
	return u.client.ListUsers(ctx, &userPb.EmptyRequest{})
}

func (u *UserUsecase) SuspendUser(ctx context.Context, userID string) (*userPb.StatusResponse, error) {
	return u.client.SuspendUser(ctx, &userPb.UserIdRequest{UserId: userID})
}
