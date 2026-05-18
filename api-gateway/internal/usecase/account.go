package usecase

import (
	"context"

	accountPb "BankingService/pb/account"
)

type AccountUsecase struct {
	client accountPb.AccountServiceClient
}

func NewAccountUsecase(client accountPb.AccountServiceClient) *AccountUsecase {
	return &AccountUsecase{client: client}
}

func (a *AccountUsecase) CreateAccount(ctx context.Context, userID, currency string) (*accountPb.AccountResponse, error) {
	return a.client.CreateAccount(ctx, &accountPb.CreateAccountRequest{UserId: userID, Currency: currency})
}

func (a *AccountUsecase) GetAccountDetails(ctx context.Context, accountID string) (*accountPb.AccountResponse, error) {
	return a.client.GetAccountDetails(ctx, &accountPb.AccountIdRequest{AccountId: accountID})
}

func (a *AccountUsecase) ListUserAccounts(ctx context.Context, userID string) (*accountPb.AccountListResponse, error) {
	return a.client.ListUserAccounts(ctx, &accountPb.UserIdRequest{UserId: userID})
}

func (a *AccountUsecase) UpdateAccountStatus(ctx context.Context, accountID, status string) (*accountPb.StatusResponse, error) {
	return a.client.UpdateAccountStatus(ctx, &accountPb.UpdateStatusRequest{AccountId: accountID, Status: status})
}

func (a *AccountUsecase) CloseAccount(ctx context.Context, accountID string) (*accountPb.StatusResponse, error) {
	return a.client.CloseAccount(ctx, &accountPb.AccountIdRequest{AccountId: accountID})
}

func (a *AccountUsecase) FreezeAccount(ctx context.Context, accountID string) (*accountPb.StatusResponse, error) {
	return a.client.FreezeAccount(ctx, &accountPb.AccountIdRequest{AccountId: accountID})
}

func (a *AccountUsecase) IssueCard(ctx context.Context, accountID, cardType string) (*accountPb.CardResponse, error) {
	return a.client.IssueCard(ctx, &accountPb.IssueCardRequest{AccountId: accountID, CardType: cardType})
}

func (a *AccountUsecase) GetCardDetails(ctx context.Context, cardID string) (*accountPb.CardResponse, error) {
	return a.client.GetCardDetails(ctx, &accountPb.CardIdRequest{CardId: cardID})
}

func (a *AccountUsecase) BlockCard(ctx context.Context, cardID string) (*accountPb.StatusResponse, error) {
	return a.client.BlockCard(ctx, &accountPb.CardIdRequest{CardId: cardID})
}

func (a *AccountUsecase) SetCardLimit(ctx context.Context, cardID string, limit float64) (*accountPb.StatusResponse, error) {
	return a.client.SetCardLimit(ctx, &accountPb.SetLimitRequest{CardId: cardID, Limit: limit})
}

func (a *AccountUsecase) UpdateAccountTier(ctx context.Context, accountID, tier string) (*accountPb.StatusResponse, error) {
	return a.client.UpdateAccountTier(ctx, &accountPb.UpdateTierRequest{AccountId: accountID, Tier: tier})
}

func (a *AccountUsecase) GetAccountLimits(ctx context.Context, accountID string) (*accountPb.LimitsResponse, error) {
	return a.client.GetAccountLimits(ctx, &accountPb.AccountIdRequest{AccountId: accountID})
}
