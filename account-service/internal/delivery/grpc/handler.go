package grpc

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"strings"

	"BankingService/account-service/internal/repository"
	pb "BankingService/pb/account"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AccountHandler struct {
	pb.UnimplementedAccountServiceServer
	repo  *repository.AccountRepository
	cache *repository.RedisCache
}

func NewAccountHandler(repo *repository.AccountRepository, cache *repository.RedisCache) *AccountHandler {
	return &AccountHandler{repo: repo, cache: cache}
}

func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	_ = ctx

	accountID := uuid.New().String()

	err := h.repo.SaveAccount(accountID, req.UserId, req.Currency)
	if err != nil {
		log.Printf("Error saving to DB: %v", err)
		return nil, err
	}

	log.Printf("New account created: %s for user: %s", accountID, req.UserId)

	return &pb.AccountResponse{
		AccountId: accountID,
		UserId:    req.UserId,
		Balance:   0.0,
		Status:    "ACTIVE",
	}, nil
}

func (h *AccountHandler) GetAccountDetails(ctx context.Context, req *pb.AccountIdRequest) (*pb.AccountResponse, error) {
	accountID := req.AccountId

	cachedData, err := h.cache.GetFromCache(ctx, accountID)
	if err == nil {
		log.Printf("[REDIS] Returned account %s from cache!", accountID)
		var response pb.AccountResponse
		_ = json.Unmarshal([]byte(cachedData), &response)
		return &response, nil
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("ERROR Redis: %v", err)
	}

	log.Printf("[POSTGRES] Searching for account %s in the database...", accountID)
	userID, _, balance, status, err := h.repo.GetAccountByID(accountID)
	if err != nil {
		log.Printf("The account was not found in the database: %v", err)
		return nil, err
	}

	response := &pb.AccountResponse{
		AccountId: accountID,
		UserId:    userID,
		Balance:   balance,
		Status:    status,
	}

	jsonData, _ := json.Marshal(response)
	_ = h.cache.SaveToCache(ctx, accountID, string(jsonData))
	return response, nil
}

func (h *AccountHandler) ListUserAccounts(ctx context.Context, req *pb.UserIdRequest) (*pb.AccountListResponse, error) {
	_ = ctx

	accounts, err := h.repo.ListAccountsByUser(req.UserId)
	if err != nil {
		return nil, err
	}

	resp := &pb.AccountListResponse{}
	for _, acc := range accounts {
		resp.Accounts = append(resp.Accounts, &pb.AccountResponse{
			AccountId: acc.AccountID,
			UserId:    acc.UserID,
			Balance:   acc.Balance,
			Status:    acc.Status,
		})
	}

	return resp, nil
}

func (h *AccountHandler) UpdateAccountStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.StatusResponse, error) {
	if err := h.repo.UpdateAccountStatus(req.AccountId, req.Status); err != nil {
		return &pb.StatusResponse{Success: false, Message: "update_failed"}, err
	}
	_ = h.cache.DeleteFromCache(ctx, req.AccountId)
	return &pb.StatusResponse{Success: true, Message: "updated"}, nil
}

func (h *AccountHandler) CloseAccount(ctx context.Context, req *pb.AccountIdRequest) (*pb.StatusResponse, error) {
	if err := h.repo.UpdateAccountStatus(req.AccountId, "CLOSED"); err != nil {
		return &pb.StatusResponse{Success: false, Message: "close_failed"}, err
	}
	_ = h.cache.DeleteFromCache(ctx, req.AccountId)
	return &pb.StatusResponse{Success: true, Message: "closed"}, nil
}

func (h *AccountHandler) IssueCard(ctx context.Context, req *pb.IssueCardRequest) (*pb.CardResponse, error) {
	_ = ctx

	cardID := uuid.New().String()
	cardNumber, err := generateCardNumber()
	if err != nil {
		return nil, err
	}

	if err := h.repo.CreateCard(cardID, req.AccountId, req.CardType, cardNumber); err != nil {
		return nil, err
	}

	return &pb.CardResponse{
		CardId:    cardID,
		AccountId: req.AccountId,
		Number:    cardNumber,
		Status:    "ACTIVE",
	}, nil
}

func (h *AccountHandler) BlockCard(ctx context.Context, req *pb.CardIdRequest) (*pb.StatusResponse, error) {
	if err := h.repo.BlockCard(req.CardId); err != nil {
		return &pb.StatusResponse{Success: false, Message: "block_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "blocked"}, nil
}

func (h *AccountHandler) GetCardDetails(ctx context.Context, req *pb.CardIdRequest) (*pb.CardResponse, error) {
	_ = ctx

	card, err := h.repo.GetCardByID(req.CardId)
	if err != nil {
		return nil, err
	}

	return &pb.CardResponse{
		CardId:    card.CardID,
		AccountId: card.AccountID,
		Number:    card.Number,
		Status:    card.Status,
	}, nil
}

func (h *AccountHandler) SetCardLimit(ctx context.Context, req *pb.SetLimitRequest) (*pb.StatusResponse, error) {
	if err := h.repo.SetCardLimit(req.CardId, req.Limit); err != nil {
		return &pb.StatusResponse{Success: false, Message: "limit_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "limit_updated"}, nil
}

func (h *AccountHandler) UpdateAccountTier(ctx context.Context, req *pb.UpdateTierRequest) (*pb.StatusResponse, error) {
	if err := h.repo.UpdateAccountTier(req.AccountId, req.Tier); err != nil {
		return &pb.StatusResponse{Success: false, Message: "tier_failed"}, err
	}
	_ = h.cache.DeleteFromCache(ctx, req.AccountId)
	return &pb.StatusResponse{Success: true, Message: "tier_updated"}, nil
}

func (h *AccountHandler) GetAccountLimits(ctx context.Context, req *pb.AccountIdRequest) (*pb.LimitsResponse, error) {
	_ = ctx

	daily, monthly, err := h.repo.GetAccountLimits(req.AccountId)
	if err != nil {
		return nil, err
	}

	return &pb.LimitsResponse{DailyLimit: daily, MonthlyLimit: monthly}, nil
}

func (h *AccountHandler) FreezeAccount(ctx context.Context, req *pb.AccountIdRequest) (*pb.StatusResponse, error) {
	if err := h.repo.UpdateAccountStatus(req.AccountId, "FROZEN"); err != nil {
		return &pb.StatusResponse{Success: false, Message: "freeze_failed"}, err
	}
	_ = h.cache.DeleteFromCache(ctx, req.AccountId)
	return &pb.StatusResponse{Success: true, Message: "frozen"}, nil
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
