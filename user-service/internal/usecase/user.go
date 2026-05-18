package usecase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"BankingService/user-service/internal/email"
	"BankingService/user-service/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenVerifyEmail   = "VERIFY_EMAIL"
	tokenResetPassword = "RESET_PASSWORD"
)

type UserUsecase struct {
	repo   UserRepository
	sender email.Sender
}

func NewUserUsecase(repo UserRepository, sender email.Sender) *UserUsecase {
	if sender == nil {
		sender = email.NewNoopSender()
	}
	return &UserUsecase{repo: repo, sender: sender}
}

func (u *UserUsecase) CreateUser(_ context.Context, name, emailAddr, password string) (string, error) {
	if password == "" {
		return "", errors.New("password is required")
	}

	userID := uuid.New().String()
	passwordHash, err := hashPassword(password)
	if err != nil {
		return "", err
	}

	if err := u.repo.SaveUser(userID, name, emailAddr, passwordHash); err != nil {
		return "", err
	}

	verifyToken, err := generateToken()
	if err != nil {
		return "", err
	}
	if err := u.repo.CreateToken(verifyToken, userID, tokenVerifyEmail, time.Now().Add(24*time.Hour)); err != nil {
		return "", err
	}

	subject := "Verify your email"
	body := fmt.Sprintf("Hello,\r\n\r\nYour verification token:\r\n%s\r\n\r\nIf you did not request this, ignore this email.\r\n", verifyToken)
	if err := u.sender.Send(emailAddr, subject, body); err != nil {
		return "", err
	}

	return userID, nil
}

func (u *UserUsecase) UpdateUserProfile(_ context.Context, userID, name string) (repository.User, error) {
	if err := u.repo.UpdateUserName(userID, name); err != nil {
		return repository.User{}, err
	}
	return u.repo.GetUserByID(userID)
}

func (u *UserUsecase) GetUserByID(_ context.Context, userID string) (repository.User, error) {
	return u.repo.GetUserByID(userID)
}

func (u *UserUsecase) DeleteUser(_ context.Context, userID string) error {
	return u.repo.DeleteUser(userID)
}

func (u *UserUsecase) ChangePassword(_ context.Context, userID, oldPassword, newPassword string) (bool, string, error) {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		return false, "", err
	}
	if !user.PasswordHash.Valid {
		return false, "password_not_set", nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(oldPassword)); err != nil {
		return false, "invalid_password", nil
	}

	newHash, err := hashPassword(newPassword)
	if err != nil {
		return false, "", err
	}

	if err := u.repo.UpdatePasswordHash(userID, newHash); err != nil {
		return false, "update_failed", err
	}

	return true, "password_changed", nil
}

func (u *UserUsecase) VerifyEmail(_ context.Context, token string) (bool, string, error) {
	t, err := u.repo.GetValidToken(token, tokenVerifyEmail)
	if err != nil {
		return false, "invalid_token", nil
	}

	if err := u.repo.SetEmailVerified(t.UserID, true); err != nil {
		return false, "verify_failed", err
	}

	_ = u.repo.MarkTokenUsed(t.Token)
	return true, "email_verified", nil
}

func (u *UserUsecase) SendPasswordReset(_ context.Context, emailAddr string) (bool, string, error) {
	user, err := u.repo.GetUserByEmail(emailAddr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, "reset_sent", nil
		}
		return false, "", err
	}

	token, err := generateToken()
	if err != nil {
		return false, "", err
	}

	expiresAt := time.Now().Add(30 * time.Minute)
	if err := u.repo.CreateToken(token, user.UserID, tokenResetPassword, expiresAt); err != nil {
		return false, "", err
	}

	subject := "Password reset"
	body := fmt.Sprintf("Hello,\r\n\r\nYour password reset token:\r\n%s\r\n\r\nIf you did not request this, ignore this email.\r\n", token)
	if err := u.sender.Send(emailAddr, subject, body); err != nil {
		return false, "", err
	}

	return true, "reset_sent", nil
}

func (u *UserUsecase) ResetPassword(_ context.Context, token, newPassword string) (bool, string, error) {
	t, err := u.repo.GetValidToken(token, tokenResetPassword)
	if err != nil {
		return false, "invalid_token", nil
	}

	newHash, err := hashPassword(newPassword)
	if err != nil {
		return false, "", err
	}

	if err := u.repo.UpdatePasswordHash(t.UserID, newHash); err != nil {
		return false, "update_failed", err
	}

	_ = u.repo.MarkTokenUsed(t.Token)
	return true, "password_reset", nil
}

func (u *UserUsecase) UpdateKYCStatus(_ context.Context, userID, status string) (bool, string, error) {
	if err := u.repo.UpdateKYCStatus(userID, status); err != nil {
		return false, "kyc_update_failed", err
	}
	return true, "kyc_updated", nil
}

func (u *UserUsecase) GetKYCStatus(_ context.Context, userID string) (string, error) {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	if user.KYCStatus.Valid {
		return user.KYCStatus.String, nil
	}
	return "", nil
}

func (u *UserUsecase) ListUsers(_ context.Context) ([]repository.User, error) {
	return u.repo.ListUsers()
}

func (u *UserUsecase) SuspendUser(_ context.Context, userID string) (bool, string, error) {
	if err := u.repo.SetSuspended(userID, true); err != nil {
		return false, "suspend_failed", err
	}
	return true, "suspended", nil
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
