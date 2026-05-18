package grpc

import (
	"context"

	pb "BankingService/pb/transaction"
	"BankingService/transaction-service/internal/usecase"
)

type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	uc *usecase.TransactionUsecase
}

func NewTransactionHandler(uc *usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

func (h *TransactionHandler) ProcessDeposit(ctx context.Context, req *pb.DepositRequest) (*pb.TransactionResponse, error) {
	txID, err := h.uc.ProcessDeposit(ctx, req.AccountId, req.Amount)
	if err != nil {
		return nil, err
	}
	return &pb.TransactionResponse{TransactionId: txID, Status: "SUCCESS"}, nil
}

func (h *TransactionHandler) ProcessWithdrawal(ctx context.Context, req *pb.WithdrawalRequest) (*pb.TransactionResponse, error) {
	txID, err := h.uc.ProcessWithdrawal(ctx, req.AccountId, req.Amount)
	if err != nil {
		return nil, err
	}
	return &pb.TransactionResponse{TransactionId: txID, Status: "SUCCESS"}, nil
}

func (h *TransactionHandler) InitiateTransfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransactionResponse, error) {
	txID, status, err := h.uc.InitiateTransfer(ctx, req.FromAccount, req.ToAccount, req.Amount)
	if err != nil {
		return nil, err
	}
	return &pb.TransactionResponse{TransactionId: txID, Status: status}, nil
}

func (h *TransactionHandler) GetTransactionStatus(ctx context.Context, req *pb.TransactionIdRequest) (*pb.StatusResponse, error) {
	status, err := h.uc.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}
	return &pb.StatusResponse{Status: status}, nil
}

func (h *TransactionHandler) GetTransactionHistory(ctx context.Context, req *pb.AccountIdRequest) (*pb.HistoryResponse, error) {
	transactions, err := h.uc.GetTransactionHistory(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}

	resp := &pb.HistoryResponse{}
	for _, tx := range transactions {
		resp.Transactions = append(resp.Transactions, &pb.TransactionResponse{TransactionId: tx.TransactionID, Status: tx.Status})
	}
	return resp, nil
}

func (h *TransactionHandler) ListPendingTransactions(ctx context.Context, _ *pb.EmptyRequest) (*pb.HistoryResponse, error) {
	transactions, err := h.uc.ListPendingTransactions(ctx)
	if err != nil {
		return nil, err
	}

	resp := &pb.HistoryResponse{}
	for _, tx := range transactions {
		resp.Transactions = append(resp.Transactions, &pb.TransactionResponse{TransactionId: tx.TransactionID, Status: tx.Status})
	}
	return resp, nil
}

func (h *TransactionHandler) ApproveTransaction(ctx context.Context, req *pb.TransactionIdRequest) (*pb.StatusResponse, error) {
	status, err := h.uc.ApproveTransaction(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}
	return &pb.StatusResponse{Status: status}, nil
}

func (h *TransactionHandler) RejectTransaction(ctx context.Context, req *pb.TransactionIdRequest) (*pb.StatusResponse, error) {
	status, err := h.uc.RejectTransaction(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}
	return &pb.StatusResponse{Status: status}, nil
}

func (h *TransactionHandler) ReverseTransaction(ctx context.Context, req *pb.TransactionIdRequest) (*pb.TransactionResponse, error) {
	reversalID, status, err := h.uc.ReverseTransaction(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}
	return &pb.TransactionResponse{TransactionId: reversalID, Status: status}, nil
}

func (h *TransactionHandler) GenerateStatement(ctx context.Context, req *pb.StatementRequest) (*pb.StatementResponse, error) {
	filePath, err := h.uc.GenerateStatement(ctx, req.AccountId, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	return &pb.StatementResponse{FileUrl: filePath}, nil
}

func (h *TransactionHandler) CalculateTransferFee(ctx context.Context, req *pb.FeeRequest) (*pb.FeeResponse, error) {
	fee := h.uc.CalculateTransferFee(ctx, req.Amount, req.TransferType)
	return &pb.FeeResponse{FeeAmount: fee}, nil
}

func (h *TransactionHandler) GetDailyTransactionVolume(ctx context.Context, _ *pb.EmptyRequest) (*pb.VolumeResponse, error) {
	total, err := h.uc.GetDailyTransactionVolume(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.VolumeResponse{TotalVolume: total}, nil
}
