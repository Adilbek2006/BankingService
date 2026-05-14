package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"log"

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
