package grpc

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	accountPb "BankingService/pb/account"
	pb "BankingService/pb/transaction"
	"BankingService/transaction-service/internal/broker"
	"BankingService/transaction-service/internal/repository"

	"github.com/google/uuid"
)

type TransactionHandler struct {
	pb.UnimplementedTransactionServiceServer
	repo          *repository.TransactionRepository
	broker        *broker.RabbitMQ
	accountClient accountPb.AccountServiceClient
}

func NewTransactionHandler(repo *repository.TransactionRepository, broker *broker.RabbitMQ, ac accountPb.AccountServiceClient) *TransactionHandler {
	return &TransactionHandler{
		repo:          repo,
		broker:        broker,
		accountClient: ac,
	}
}

const (
	statusPending  = "PENDING"
	statusSuccess  = "SUCCESS"
	statusRejected = "REJECTED"
	statusReversed = "REVERSED"

	typeDeposit    = "DEPOSIT"
	typeWithdrawal = "WITHDRAWAL"
	typeTransfer   = "TRANSFER"
	typeReversal   = "REVERSAL"
)

func (h *TransactionHandler) ProcessDeposit(ctx context.Context, req *pb.DepositRequest) (*pb.TransactionResponse, error) {
	_ = ctx

	transactionID := uuid.New().String()

	err := h.repo.SaveTransaction(transactionID, req.AccountId, typeDeposit, req.Amount, statusSuccess)
	if err != nil {
		log.Printf("Failed to save transaction: %v", err)
		return nil, err
	}

	err = h.broker.PublishDeposit(req.AccountId, req.Amount)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
	}

	log.Printf("Deposit processed: %s | Account: %s | Amount: %f", transactionID, req.AccountId, req.Amount)

	return &pb.TransactionResponse{
		TransactionId: transactionID,
		Status:        statusSuccess,
	}, nil
}

func (h *TransactionHandler) ProcessWithdrawal(ctx context.Context, req *pb.WithdrawalRequest) (*pb.TransactionResponse, error) {
	accDetails, err := h.accountClient.GetAccountDetails(ctx, &accountPb.AccountIdRequest{
		AccountId: req.AccountId,
	})
	if err != nil {
		return nil, err
	}

	if accDetails.Balance < req.Amount {
		return nil, errors.New("insufficient funds")
	}

	transactionID := uuid.New().String()

	if err := h.repo.SaveTransaction(transactionID, req.AccountId, typeWithdrawal, req.Amount, statusSuccess); err != nil {
		return nil, err
	}

	if err := h.broker.PublishWithdrawal(req.AccountId, req.Amount); err != nil {
		log.Printf("Failed to publish withdrawal: %v", err)
	}

	return &pb.TransactionResponse{TransactionId: transactionID, Status: statusSuccess}, nil
}

func (h *TransactionHandler) InitiateTransfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransactionResponse, error) {
	accDetails, err := h.accountClient.GetAccountDetails(ctx, &accountPb.AccountIdRequest{
		AccountId: req.FromAccount,
	})
	if err != nil {
		return nil, err
	}

	if accDetails.Balance < req.Amount {
		return nil, errors.New("insufficient funds")
	}

	txID := uuid.New().String()

	err = h.repo.SaveTransfer(txID, req.FromAccount, req.ToAccount, req.Amount, statusPending)
	if err != nil {
		return nil, err
	}

	return &pb.TransactionResponse{
		TransactionId: txID,
		Status:        statusPending,
	}, nil
}

func (h *TransactionHandler) GetTransactionStatus(ctx context.Context, req *pb.TransactionIdRequest) (*pb.StatusResponse, error) {
	_ = ctx

	tx, err := h.repo.GetTransactionByID(req.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.StatusResponse{Status: tx.Status}, nil
}

func (h *TransactionHandler) GetTransactionHistory(ctx context.Context, req *pb.AccountIdRequest) (*pb.HistoryResponse, error) {
	_ = ctx

	transactions, err := h.repo.ListTransactionsByAccount(req.AccountId)
	if err != nil {
		return nil, err
	}

	resp := &pb.HistoryResponse{}
	for _, tx := range transactions {
		resp.Transactions = append(resp.Transactions, &pb.TransactionResponse{
			TransactionId: tx.TransactionID,
			Status:        tx.Status,
		})
	}

	return resp, nil
}

func (h *TransactionHandler) ListPendingTransactions(ctx context.Context, _ *pb.EmptyRequest) (*pb.HistoryResponse, error) {
	_ = ctx

	transactions, err := h.repo.ListPendingTransactions()
	if err != nil {
		return nil, err
	}

	resp := &pb.HistoryResponse{}
	for _, tx := range transactions {
		resp.Transactions = append(resp.Transactions, &pb.TransactionResponse{
			TransactionId: tx.TransactionID,
			Status:        tx.Status,
		})
	}

	return resp, nil
}

func (h *TransactionHandler) ApproveTransaction(ctx context.Context, req *pb.TransactionIdRequest) (*pb.StatusResponse, error) {
	tx, err := h.repo.GetTransactionByID(req.TransactionId)
	if err != nil {
		return nil, err
	}

	if tx.Status != statusPending {
		return &pb.StatusResponse{Status: tx.Status}, nil
	}

	if tx.Type != typeTransfer {
		return nil, errors.New("only transfer transactions can be approved")
	}

	if !tx.ToAccountID.Valid {
		return nil, errors.New("missing destination account")
	}

	if err := h.broker.PublishTransfer(tx.AccountID, tx.ToAccountID.String, tx.Amount); err != nil {
		log.Printf("failed to publish transfer: %v", err)
	}

	if err := h.repo.UpdateTransactionStatus(tx.TransactionID, statusSuccess); err != nil {
		return nil, err
	}

	return &pb.StatusResponse{Status: statusSuccess}, nil
}

func (h *TransactionHandler) RejectTransaction(ctx context.Context, req *pb.TransactionIdRequest) (*pb.StatusResponse, error) {
	tx, err := h.repo.GetTransactionByID(req.TransactionId)
	if err != nil {
		return nil, err
	}

	if tx.Status != statusPending {
		return &pb.StatusResponse{Status: tx.Status}, nil
	}

	if err := h.repo.UpdateTransactionStatus(tx.TransactionID, statusRejected); err != nil {
		return nil, err
	}

	return &pb.StatusResponse{Status: statusRejected}, nil
}

func (h *TransactionHandler) ReverseTransaction(ctx context.Context, req *pb.TransactionIdRequest) (*pb.TransactionResponse, error) {
	tx, err := h.repo.GetTransactionByID(req.TransactionId)
	if err != nil {
		return nil, err
	}

	if tx.Status == statusReversed {
		return &pb.TransactionResponse{TransactionId: req.TransactionId, Status: statusReversed}, nil
	}

	reversalID := uuid.New().String()

	switch tx.Type {
	case typeTransfer:
		if !tx.ToAccountID.Valid {
			return nil, errors.New("missing destination account")
		}
		if err := h.broker.PublishTransfer(tx.ToAccountID.String, tx.AccountID, tx.Amount); err != nil {
			log.Printf("failed to publish reversal transfer: %v", err)
		}
		if err := h.repo.InsertReversal(reversalID, tx.ToAccountID.String, sql.NullString{String: tx.AccountID, Valid: true}, tx.Amount, tx.TransactionID); err != nil {
			return nil, err
		}
	case typeDeposit:
		if err := h.broker.PublishWithdrawal(tx.AccountID, tx.Amount); err != nil {
			log.Printf("failed to publish reversal deposit: %v", err)
		}
		if err := h.repo.InsertReversal(reversalID, tx.AccountID, sql.NullString{}, tx.Amount, tx.TransactionID); err != nil {
			return nil, err
		}
	case typeWithdrawal:
		if err := h.broker.PublishDeposit(tx.AccountID, tx.Amount); err != nil {
			log.Printf("failed to publish reversal withdrawal: %v", err)
		}
		if err := h.repo.InsertReversal(reversalID, tx.AccountID, sql.NullString{}, tx.Amount, tx.TransactionID); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported transaction type for reversal")
	}

	if err := h.repo.MarkTransactionReversed(tx.TransactionID); err != nil {
		return nil, err
	}

	return &pb.TransactionResponse{TransactionId: reversalID, Status: statusReversed}, nil
}

func (h *TransactionHandler) GenerateStatement(ctx context.Context, req *pb.StatementRequest) (*pb.StatementResponse, error) {
	_ = ctx

	start, end, err := parseStatementRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	transactions, err := h.repo.ListTransactionsByAccountInRange(req.AccountId, start, end)
	if err != nil {
		return nil, err
	}

	filePath, err := writeStatementCSV(req.AccountId, start, end, transactions)
	if err != nil {
		return nil, err
	}

	return &pb.StatementResponse{FileUrl: filePath}, nil
}

func (h *TransactionHandler) CalculateTransferFee(ctx context.Context, req *pb.FeeRequest) (*pb.FeeResponse, error) {
	_ = ctx

	feeRate := 0.005
	switch req.TransferType {
	case "INTERNAL":
		feeRate = 0
	case "EXTERNAL":
		feeRate = 0.01
	}

	fee := req.Amount * feeRate
	return &pb.FeeResponse{FeeAmount: fee}, nil
}

func (h *TransactionHandler) GetDailyTransactionVolume(ctx context.Context, _ *pb.EmptyRequest) (*pb.VolumeResponse, error) {
	_ = ctx

	total, err := h.repo.SumDailyVolume()
	if err != nil {
		return nil, err
	}

	return &pb.VolumeResponse{TotalVolume: total}, nil
}

func parseStatementRange(startDate, endDate string) (time.Time, time.Time, error) {
	now := time.Now()
	if startDate == "" && endDate == "" {
		return now.AddDate(0, 0, -30), now, nil
	}

	if endDate == "" {
		endDate = now.Format("2006-01-02")
	}

	if startDate == "" {
		startDate = now.AddDate(0, 0, -30).Format("2006-01-02")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	return start, end, nil
}

func writeStatementCSV(accountID string, start, end time.Time, transactions []repository.Transaction) (string, error) {
	filename := fmt.Sprintf("statement_%s_%s_%s.csv", accountID, start.Format("20060102"), end.Format("20060102"))
	filePath := filepath.Join(os.TempDir(), filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"transaction_id", "account_id", "to_account_id", "amount", "type", "status", "created_at"}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	for _, tx := range transactions {
		toAccount := ""
		if tx.ToAccountID.Valid {
			toAccount = tx.ToAccountID.String
		}

		row := []string{
			tx.TransactionID,
			tx.AccountID,
			toAccount,
			strconv.FormatFloat(tx.Amount, 'f', 2, 64),
			tx.Type,
			tx.Status,
			tx.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	return filePath, nil
}
