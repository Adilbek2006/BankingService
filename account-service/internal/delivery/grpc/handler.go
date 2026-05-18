package grpc

import (
	"context"

	"BankingService/account-service/internal/usecase"
	pb "BankingService/pb/account"
)

type AccountHandler struct {
	pb.UnimplementedAccountServiceServer
	uc *usecase.AccountUsecase
}

func NewAccountHandler(uc *usecase.AccountUsecase) *AccountHandler {
	return &AccountHandler{uc: uc}
}

func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	accountID, err := h.uc.CreateAccount(ctx, req.UserId, req.Currency)
	if err != nil {
		return nil, err
	}

	return &pb.AccountResponse{
		AccountId: accountID,
		UserId:    req.UserId,
		Balance:   0.0,
		Status:    "ACTIVE",
	}, nil
}

func (h *AccountHandler) GetAccountDetails(ctx context.Context, req *pb.AccountIdRequest) (*pb.AccountResponse, error) {
	acc, err := h.uc.GetAccountDetails(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}

	return &pb.AccountResponse{
		AccountId: acc.AccountID,
		UserId:    acc.UserID,
		Balance:   acc.Balance,
		Status:    acc.Status,
	}, nil
}

func (h *AccountHandler) ListUserAccounts(ctx context.Context, req *pb.UserIdRequest) (*pb.AccountListResponse, error) {
	accounts, err := h.uc.ListUserAccounts(ctx, req.UserId)
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
	if err := h.uc.UpdateAccountStatus(ctx, req.AccountId, req.Status); err != nil {
		return &pb.StatusResponse{Success: false, Message: "update_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "updated"}, nil
}

func (h *AccountHandler) CloseAccount(ctx context.Context, req *pb.AccountIdRequest) (*pb.StatusResponse, error) {
	if err := h.uc.CloseAccount(ctx, req.AccountId); err != nil {
		return &pb.StatusResponse{Success: false, Message: "close_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "closed"}, nil
}

func (h *AccountHandler) IssueCard(ctx context.Context, req *pb.IssueCardRequest) (*pb.CardResponse, error) {
	card, err := h.uc.IssueCard(ctx, req.AccountId, req.CardType)
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

func (h *AccountHandler) BlockCard(ctx context.Context, req *pb.CardIdRequest) (*pb.StatusResponse, error) {
	if err := h.uc.BlockCard(ctx, req.CardId); err != nil {
		return &pb.StatusResponse{Success: false, Message: "block_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "blocked"}, nil
}

func (h *AccountHandler) GetCardDetails(ctx context.Context, req *pb.CardIdRequest) (*pb.CardResponse, error) {
	card, err := h.uc.GetCardDetails(ctx, req.CardId)
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
	if err := h.uc.SetCardLimit(ctx, req.CardId, req.Limit); err != nil {
		return &pb.StatusResponse{Success: false, Message: "limit_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "limit_updated"}, nil
}

func (h *AccountHandler) UpdateAccountTier(ctx context.Context, req *pb.UpdateTierRequest) (*pb.StatusResponse, error) {
	if err := h.uc.UpdateAccountTier(ctx, req.AccountId, req.Tier); err != nil {
		return &pb.StatusResponse{Success: false, Message: "tier_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "tier_updated"}, nil
}

func (h *AccountHandler) GetAccountLimits(ctx context.Context, req *pb.AccountIdRequest) (*pb.LimitsResponse, error) {
	daily, monthly, err := h.uc.GetAccountLimits(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}
	return &pb.LimitsResponse{DailyLimit: daily, MonthlyLimit: monthly}, nil
}

func (h *AccountHandler) FreezeAccount(ctx context.Context, req *pb.AccountIdRequest) (*pb.StatusResponse, error) {
	if err := h.uc.FreezeAccount(ctx, req.AccountId); err != nil {
		return &pb.StatusResponse{Success: false, Message: "freeze_failed"}, err
	}
	return &pb.StatusResponse{Success: true, Message: "frozen"}, nil
}
