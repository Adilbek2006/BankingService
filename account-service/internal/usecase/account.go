package usecase

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"BankingService/account-service/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AccountUsecase struct {
	repo  AccountRepository
	cache AccountCache
}

func NewAccountUsecase(repo AccountRepository, cache AccountCache) *AccountUsecase {
	return &AccountUsecase{repo: repo, cache: cache}
}

func (u *AccountUsecase) CreateAccount(_ context.Context, userID, currency string) (string, error) {
	accountID := uuid.New().String()
	if err := u.repo.SaveAccount(accountID, userID, currency); err != nil {
		return "", err
	}
	return accountID, nil
}

func (u *AccountUsecase) GetAccountDetails(ctx context.Context, accountID string) (repository.Account, error) {
	cachedData, err := u.cache.GetFromCache(ctx, accountID)
	if err == nil {
		var response repository.Account
		_ = json.Unmarshal([]byte(cachedData), &response)
		return response, nil
	} else if !errors.Is(err, redis.Nil) {
		return repository.Account{}, err
	}

	userID, currency, balance, status, err := u.repo.GetAccountByID(accountID)
	if err != nil {
		return repository.Account{}, err
	}

	acc := repository.Account{
		AccountID: accountID,
		UserID:    userID,
		Currency:  currency,
		Balance:   balance,
		Status:    status,
	}

	jsonData, _ := json.Marshal(acc)
	_ = u.cache.SaveToCache(ctx, accountID, string(jsonData))
	return acc, nil
}

func (u *AccountUsecase) ListUserAccounts(_ context.Context, userID string) ([]repository.Account, error) {
	return u.repo.ListAccountsByUser(userID)
}

func (u *AccountUsecase) UpdateAccountStatus(ctx context.Context, accountID, status string) error {
	if err := u.repo.UpdateAccountStatus(accountID, status); err != nil {
		return err
	}
	_ = u.cache.DeleteFromCache(ctx, accountID)
	return nil
}

func (u *AccountUsecase) CloseAccount(ctx context.Context, accountID string) error {
	if err := u.repo.UpdateAccountStatus(accountID, "CLOSED"); err != nil {
		return err
	}
	_ = u.cache.DeleteFromCache(ctx, accountID)
	return nil
}

func (u *AccountUsecase) IssueCard(_ context.Context, accountID, cardType string) (repository.Card, error) {
	cardID := uuid.New().String()
	cardNumber, err := generateCardNumber()
	if err != nil {
		return repository.Card{}, err
	}

	if err := u.repo.CreateCard(cardID, accountID, cardType, cardNumber); err != nil {
		return repository.Card{}, err
	}

	return repository.Card{
		CardID:    cardID,
		AccountID: accountID,
		CardType:  cardType,
		Number:    cardNumber,
		Status:    "ACTIVE",
	}, nil
}

func (u *AccountUsecase) BlockCard(_ context.Context, cardID string) error {
	return u.repo.BlockCard(cardID)
}

func (u *AccountUsecase) GetCardDetails(_ context.Context, cardID string) (repository.Card, error) {
	return u.repo.GetCardByID(cardID)
}

func (u *AccountUsecase) SetCardLimit(_ context.Context, cardID string, limit float64) error {
	return u.repo.SetCardLimit(cardID, limit)
}

func (u *AccountUsecase) UpdateAccountTier(ctx context.Context, accountID, tier string) error {
	if err := u.repo.UpdateAccountTier(accountID, tier); err != nil {
		return err
	}
	_ = u.cache.DeleteFromCache(ctx, accountID)
	return nil
}

func (u *AccountUsecase) GetAccountLimits(_ context.Context, accountID string) (float64, float64, error) {
	return u.repo.GetAccountLimits(accountID)
}

func (u *AccountUsecase) FreezeAccount(ctx context.Context, accountID string) error {
	if err := u.repo.UpdateAccountStatus(accountID, "FROZEN"); err != nil {
		return err
	}
	_ = u.cache.DeleteFromCache(ctx, accountID)
	return nil
}

func generateCardNumber() (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(16), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return padLeft(n.String(), 16), nil
}

func padLeft(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return strings.Repeat("0", width-len(value)) + value
}
