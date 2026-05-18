package usecase

import (
	"time"

	"BankingService/user-service/internal/repository"
)

type UserRepository interface {
	SaveUser(userID, name, email, passwordHash string) error
	GetUserByID(userID string) (repository.User, error)
	GetUserByEmail(email string) (repository.User, error)
	UpdateUserName(userID, name string) error
	UpdatePasswordHash(userID, passwordHash string) error
	UpdateKYCStatus(userID, status string) error
	SetSuspended(userID string, suspended bool) error
	SetEmailVerified(userID string, verified bool) error
	DeleteUser(userID string) error
	ListUsers() ([]repository.User, error)
	CreateToken(token, userID, tokenType string, expiresAt time.Time) error
	GetValidToken(token, tokenType string) (repository.UserToken, error)
	MarkTokenUsed(token string) error
}
