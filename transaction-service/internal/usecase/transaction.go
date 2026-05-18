package usecase

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	accountPb "BankingService/pb/account"
	"BankingService/transaction-service/internal/broker"
	"BankingService/transaction-service/internal/repository"

	"github.com/google/uuid"
)

const (
	statusPending  = "PENDING"
	statusSuccess  = "SUCCESS"
	statusRejected = "REJECTED"
	statusReversed = "REVERSED"

	typeDeposit    = "DEPOSIT"
	typeWithdrawal = "WITHDRAWAL"
	typeTransfer   = "TRANSFER"
)

type TransactionUsecase struct {
	repo          *repository.TransactionRepository
	broker        *broker.RabbitMQ
	accountClient accountPb.AccountServiceClient
}

func NewTransactionUsecase(repo *repository.TransactionRepository, broker *broker.RabbitMQ, ac accountPb.AccountServiceClient) *TransactionUsecase {
	return &TransactionUsecase{repo: repo, broker: broker, accountClient: ac}
}

func (u *TransactionUsecase) ProcessDeposit(ctx context.Context, accountID string, amount float64) (string, error) {
	transactionID := uuid.New().String()

	if err := u.repo.SaveTransaction(transactionID, accountID, typeDeposit, amount, statusSuccess); err != nil {
		return "", err
	}

	if err := u.broker.PublishDeposit(accountID, amount); err != nil {
		return "", err
	}

	return transactionID, nil
}

func (u *TransactionUsecase) ProcessWithdrawal(ctx context.Context, accountID string, amount float64) (string, error) {
	accDetails, err := u.accountClient.GetAccountDetails(ctx, &accountPb.AccountIdRequest{AccountId: accountID})
	if err != nil {
		return "", err
	}
	if accDetails.Status == "FROZEN" {
		return "", errors.New("account is frozen")
	}
	if accDetails.Balance < amount {
		return "", errors.New("insufficient funds")
	}

	transactionID := uuid.New().String()
	if err := u.repo.SaveTransaction(transactionID, accountID, typeWithdrawal, amount, statusSuccess); err != nil {
		return "", err
	}

	if err := u.broker.PublishWithdrawal(accountID, amount); err != nil {
		return "", err
	}

	return transactionID, nil
}

func (u *TransactionUsecase) InitiateTransfer(ctx context.Context, fromAccount, toAccount string, amount float64) (string, string, error) {
	accDetails, err := u.accountClient.GetAccountDetails(ctx, &accountPb.AccountIdRequest{AccountId: fromAccount})
	if err != nil {
		return "", "", err
	}
	if accDetails.Status == "FROZEN" {
		return "", "", errors.New("account is frozen")
	}
	if accDetails.Balance < amount {
		return "", "", errors.New("insufficient funds")
	}

	txID := uuid.New().String()
	if err := u.repo.SaveTransfer(txID, fromAccount, toAccount, amount, statusPending); err != nil {
		return "", "", err
	}

	return txID, statusPending, nil
}

func (u *TransactionUsecase) GetTransactionStatus(_ context.Context, id string) (string, error) {
	tx, err := u.repo.GetTransactionByID(id)
	if err != nil {
		return "", err
	}
	return tx.Status, nil
}

func (u *TransactionUsecase) GetTransactionHistory(_ context.Context, accountID string) ([]repository.Transaction, error) {
	return u.repo.ListTransactionsByAccount(accountID)
}

func (u *TransactionUsecase) ListPendingTransactions(_ context.Context) ([]repository.Transaction, error) {
	return u.repo.ListPendingTransactions()
}

func (u *TransactionUsecase) ApproveTransaction(ctx context.Context, id string) (string, error) {
	tx, err := u.repo.GetTransactionByID(id)
	if err != nil {
		return "", err
	}
	if tx.Status != statusPending {
		return tx.Status, nil
	}
	if tx.Type != typeTransfer {
		return "", errors.New("only transfer transactions can be approved")
	}
	if !tx.ToAccountID.Valid {
		return "", errors.New("missing destination account")
	}

	accDetails, err := u.accountClient.GetAccountDetails(ctx, &accountPb.AccountIdRequest{AccountId: tx.AccountID})
	if err != nil {
		return "", err
	}
	if accDetails.Status == "FROZEN" {
		return "", errors.New("account is frozen")
	}

	if err := u.broker.PublishTransfer(tx.AccountID, tx.ToAccountID.String, tx.Amount); err != nil {
		return "", err
	}

	if err := u.repo.UpdateTransactionStatus(tx.TransactionID, statusSuccess); err != nil {
		return "", err
	}
	return statusSuccess, nil
}

func (u *TransactionUsecase) RejectTransaction(_ context.Context, id string) (string, error) {
	tx, err := u.repo.GetTransactionByID(id)
	if err != nil {
		return "", err
	}
	if tx.Status != statusPending {
		return tx.Status, nil
	}
	if err := u.repo.UpdateTransactionStatus(tx.TransactionID, statusRejected); err != nil {
		return "", err
	}
	return statusRejected, nil
}

func (u *TransactionUsecase) ReverseTransaction(_ context.Context, id string) (string, string, error) {
	tx, err := u.repo.GetTransactionByID(id)
	if err != nil {
		return "", "", err
	}
	if tx.Status == statusReversed {
		return id, statusReversed, nil
	}

	reversalID := uuid.New().String()

	switch tx.Type {
	case typeTransfer:
		if !tx.ToAccountID.Valid {
			return "", "", errors.New("missing destination account")
		}
		if err := u.broker.PublishTransfer(tx.ToAccountID.String, tx.AccountID, tx.Amount); err != nil {
			return "", "", err
		}
		if err := u.repo.InsertReversal(reversalID, tx.ToAccountID.String, sql.NullString{String: tx.AccountID, Valid: true}, tx.Amount, tx.TransactionID); err != nil {
			return "", "", err
		}
	case typeDeposit:
		if err := u.broker.PublishWithdrawal(tx.AccountID, tx.Amount); err != nil {
			return "", "", err
		}
		if err := u.repo.InsertReversal(reversalID, tx.AccountID, sql.NullString{}, tx.Amount, tx.TransactionID); err != nil {
			return "", "", err
		}
	case typeWithdrawal:
		if err := u.broker.PublishDeposit(tx.AccountID, tx.Amount); err != nil {
			return "", "", err
		}
		if err := u.repo.InsertReversal(reversalID, tx.AccountID, sql.NullString{}, tx.Amount, tx.TransactionID); err != nil {
			return "", "", err
		}
	default:
		return "", "", errors.New("unsupported transaction type for reversal")
	}

	if err := u.repo.MarkTransactionReversed(tx.TransactionID); err != nil {
		return "", "", err
	}

	return reversalID, statusReversed, nil
}

func (u *TransactionUsecase) GenerateStatement(_ context.Context, accountID, startDate, endDate string) (string, error) {
	start, end, err := parseStatementRange(startDate, endDate)
	if err != nil {
		return "", err
	}

	transactions, err := u.repo.ListTransactionsByAccountInRange(accountID, start, end)
	if err != nil {
		return "", err
	}

	return writeStatementCSV(accountID, start, end, transactions)
}

func (u *TransactionUsecase) CalculateTransferFee(_ context.Context, amount float64, transferType string) float64 {
	feeRate := 0.005
	switch transferType {
	case "INTERNAL":
		feeRate = 0
	case "EXTERNAL":
		feeRate = 0.01
	}

	return amount * feeRate
}

func (u *TransactionUsecase) GetDailyTransactionVolume(_ context.Context) (float64, error) {
	return u.repo.SumDailyVolume()
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
