package usecase

import (
	"context"

	transactionPb "BankingService/pb/transaction"
)

type TransactionUsecase struct {
	client transactionPb.TransactionServiceClient
}

func NewTransactionUsecase(client transactionPb.TransactionServiceClient) *TransactionUsecase {
	return &TransactionUsecase{client: client}
}

func (t *TransactionUsecase) ProcessDeposit(ctx context.Context, accountID string, amount float64) (*transactionPb.TransactionResponse, error) {
	return t.client.ProcessDeposit(ctx, &transactionPb.DepositRequest{AccountId: accountID, Amount: amount})
}

func (t *TransactionUsecase) InitiateTransfer(ctx context.Context, fromAccount, toAccount string, amount float64) (*transactionPb.TransactionResponse, error) {
	return t.client.InitiateTransfer(ctx, &transactionPb.TransferRequest{FromAccount: fromAccount, ToAccount: toAccount, Amount: amount})
}

func (t *TransactionUsecase) ProcessWithdrawal(ctx context.Context, accountID string, amount float64) (*transactionPb.TransactionResponse, error) {
	return t.client.ProcessWithdrawal(ctx, &transactionPb.WithdrawalRequest{AccountId: accountID, Amount: amount})
}

func (t *TransactionUsecase) GetTransactionStatus(ctx context.Context, transactionID string) (*transactionPb.StatusResponse, error) {
	return t.client.GetTransactionStatus(ctx, &transactionPb.TransactionIdRequest{TransactionId: transactionID})
}

func (t *TransactionUsecase) GetTransactionHistory(ctx context.Context, accountID string) (*transactionPb.HistoryResponse, error) {
	return t.client.GetTransactionHistory(ctx, &transactionPb.AccountIdRequest{AccountId: accountID})
}

func (t *TransactionUsecase) ListPendingTransactions(ctx context.Context) (*transactionPb.HistoryResponse, error) {
	return t.client.ListPendingTransactions(ctx, &transactionPb.EmptyRequest{})
}

func (t *TransactionUsecase) ApproveTransaction(ctx context.Context, transactionID string) (*transactionPb.StatusResponse, error) {
	return t.client.ApproveTransaction(ctx, &transactionPb.TransactionIdRequest{TransactionId: transactionID})
}

func (t *TransactionUsecase) RejectTransaction(ctx context.Context, transactionID string) (*transactionPb.StatusResponse, error) {
	return t.client.RejectTransaction(ctx, &transactionPb.TransactionIdRequest{TransactionId: transactionID})
}

func (t *TransactionUsecase) ReverseTransaction(ctx context.Context, transactionID string) (*transactionPb.TransactionResponse, error) {
	return t.client.ReverseTransaction(ctx, &transactionPb.TransactionIdRequest{TransactionId: transactionID})
}

func (t *TransactionUsecase) GenerateStatement(ctx context.Context, accountID, startDate, endDate string) (*transactionPb.StatementResponse, error) {
	return t.client.GenerateStatement(ctx, &transactionPb.StatementRequest{AccountId: accountID, StartDate: startDate, EndDate: endDate})
}

func (t *TransactionUsecase) CalculateTransferFee(ctx context.Context, amount float64, transferType string) (*transactionPb.FeeResponse, error) {
	return t.client.CalculateTransferFee(ctx, &transactionPb.FeeRequest{Amount: amount, TransferType: transferType})
}

func (t *TransactionUsecase) GetDailyTransactionVolume(ctx context.Context) (*transactionPb.VolumeResponse, error) {
	return t.client.GetDailyTransactionVolume(ctx, &transactionPb.EmptyRequest{})
}
