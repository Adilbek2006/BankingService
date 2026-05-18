package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"BankingService/account-service/internal/repository"

	"github.com/redis/go-redis/v9"
)

type fakeAccountRepo struct {
	account repository.Account
	saved   []repository.Account
}

func (r *fakeAccountRepo) SaveAccount(accountID, userID, currency string) error {
	r.account = repository.Account{AccountID: accountID, UserID: userID, Currency: currency, Balance: 0, Status: "ACTIVE"}
	r.saved = append(r.saved, r.account)
	return nil
}

func (r *fakeAccountRepo) GetAccountByID(accountID string) (string, string, float64, string, error) {
	if r.account.AccountID != accountID {
		return "", "", 0, "", errors.New("not found")
	}
	return r.account.UserID, r.account.Currency, r.account.Balance, r.account.Status, nil
}

func (r *fakeAccountRepo) ListAccountsByUser(userID string) ([]repository.Account, error) {
	if r.account.UserID != userID {
		return nil, nil
	}
	return []repository.Account{r.account}, nil
}

func (r *fakeAccountRepo) UpdateAccountStatus(accountID, status string) error {
	if r.account.AccountID != accountID {
		return errors.New("not found")
	}
	r.account.Status = status
	return nil
}

func (r *fakeAccountRepo) UpdateAccountTier(accountID, tier string) error {
	if r.account.AccountID != accountID {
		return errors.New("not found")
	}
	r.account.Tier = tier
	return nil
}

func (r *fakeAccountRepo) GetAccountLimits(accountID string) (float64, float64, error) {
	if r.account.AccountID != accountID {
		return 0, 0, errors.New("not found")
	}
	return 100, 1000, nil
}

func (r *fakeAccountRepo) CreateCard(cardID, accountID, cardType, number string) error {
	return nil
}

func (r *fakeAccountRepo) GetCardByID(cardID string) (repository.Card, error) {
	return repository.Card{CardID: cardID}, nil
}

func (r *fakeAccountRepo) BlockCard(cardID string) error { return nil }

func (r *fakeAccountRepo) SetCardLimit(cardID string, limit float64) error { return nil }

type fakeAccountCache struct {
	data         map[string]string
	deleteCalled bool
}

func newFakeAccountCache() *fakeAccountCache {
	return &fakeAccountCache{data: map[string]string{}}
}

func (c *fakeAccountCache) SaveToCache(_ context.Context, accountID string, data string) error {
	c.data[accountID] = data
	return nil
}

func (c *fakeAccountCache) GetFromCache(_ context.Context, accountID string) (string, error) {
	val, ok := c.data[accountID]
	if !ok {
		return "", redis.Nil
	}
	return val, nil
}

func (c *fakeAccountCache) DeleteFromCache(_ context.Context, accountID string) error {
	delete(c.data, accountID)
	c.deleteCalled = true
	return nil
}

func TestGetAccountDetailsCacheHit(t *testing.T) {
	cache := newFakeAccountCache()
	repo := &fakeAccountRepo{}
	uc := NewAccountUsecase(repo, cache)

	acc := repository.Account{AccountID: "acc-1", UserID: "user-1", Currency: "USD", Balance: 10, Status: "ACTIVE"}
	payload, _ := json.Marshal(acc)
	cache.data["acc-1"] = string(payload)

	got, err := uc.GetAccountDetails(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccountID != "acc-1" || got.Balance != 10 {
		t.Fatalf("unexpected account: %+v", got)
	}
}

func TestGetAccountDetailsCacheMiss(t *testing.T) {
	cache := newFakeAccountCache()
	repo := &fakeAccountRepo{account: repository.Account{AccountID: "acc-1", UserID: "user-1", Currency: "USD", Balance: 10, Status: "ACTIVE"}}
	uc := NewAccountUsecase(repo, cache)

	got, err := uc.GetAccountDetails(context.Background(), "acc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccountID != "acc-1" || cache.data["acc-1"] == "" {
		t.Fatal("expected cached value after miss")
	}
}

func TestFreezeAccountClearsCache(t *testing.T) {
	cache := newFakeAccountCache()
	repo := &fakeAccountRepo{account: repository.Account{AccountID: "acc-1", Status: "ACTIVE"}}
	uc := NewAccountUsecase(repo, cache)

	if err := uc.FreezeAccount(context.Background(), "acc-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cache.deleteCalled {
		t.Fatal("expected cache delete")
	}
	if repo.account.Status != "FROZEN" {
		t.Fatalf("expected FROZEN, got %s", repo.account.Status)
	}
}
