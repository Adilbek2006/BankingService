package usecase

import (
	"context"

	"BankingService/account-service/internal/repository"
)

type AccountRepository interface {
	SaveAccount(accountID, userID, currency string) error
	GetAccountByID(accountID string) (string, string, float64, string, error)
	ListAccountsByUser(userID string) ([]repository.Account, error)
	UpdateAccountStatus(accountID, status string) error
	UpdateAccountTier(accountID, tier string) error
	GetAccountLimits(accountID string) (float64, float64, error)
	CreateCard(cardID, accountID, cardType, number string) error
	GetCardByID(cardID string) (repository.Card, error)
	BlockCard(cardID string) error
	SetCardLimit(cardID string, limit float64) error
}

type AccountCache interface {
	SaveToCache(ctx context.Context, accountID string, data string) error
	GetFromCache(ctx context.Context, accountID string) (string, error)
	DeleteFromCache(ctx context.Context, accountID string) error
}
