//go:build integration
// +build integration

package repository

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"BankingService/internal/config"

	"github.com/google/uuid"
)

func TestTransactionRepositoryIntegration(t *testing.T) {
	loadEnvForTests()

	db := openTestDB(t)
	defer db.Close()

	repo := NewTransactionRepository(db)

	accountID := uuid.New().String()
	toAccountID := uuid.New().String()

	txID := uuid.New().String()
	if err := repo.SaveTransaction(txID, accountID, "DEPOSIT", 100, "SUCCESS"); err != nil {
		t.Fatalf("SaveTransaction failed: %v", err)
	}

	transferID := uuid.New().String()
	if err := repo.SaveTransfer(transferID, accountID, toAccountID, 25, "PENDING"); err != nil {
		t.Fatalf("SaveTransfer failed: %v", err)
	}

	defer func() {
		_, _ = db.Exec("DELETE FROM transactions WHERE transaction_id IN ($1, $2)", txID, transferID)
	}()

	tx, err := repo.GetTransactionByID(txID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}
	if tx.AccountID != accountID || tx.Type != "DEPOSIT" {
		t.Fatalf("unexpected transaction: %+v", tx)
	}

	if err := repo.UpdateTransactionStatus(transferID, "SUCCESS"); err != nil {
		t.Fatalf("UpdateTransactionStatus failed: %v", err)
	}

	list, err := repo.ListTransactionsByAccount(accountID)
	if err != nil {
		t.Fatalf("ListTransactionsByAccount failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one transaction")
	}

	pending, err := repo.ListPendingTransactions()
	if err != nil {
		t.Fatalf("ListPendingTransactions failed: %v", err)
	}
	for _, p := range pending {
		if p.TransactionID == transferID {
			t.Fatal("expected transfer to be SUCCESS, not pending")
		}
	}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)
	inRange, err := repo.ListTransactionsByAccountInRange(accountID, start, end)
	if err != nil {
		t.Fatalf("ListTransactionsByAccountInRange failed: %v", err)
	}
	if len(inRange) == 0 {
		t.Fatal("expected transactions in range")
	}

	before, err := repo.SumDailyVolume()
	if err != nil {
		t.Fatalf("SumDailyVolume failed: %v", err)
	}

	volumeTxID := uuid.New().String()
	if err := repo.SaveTransaction(volumeTxID, accountID, "DEPOSIT", 10, "SUCCESS"); err != nil {
		t.Fatalf("SaveTransaction for volume failed: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM transactions WHERE transaction_id = $1", volumeTxID)
	}()

	after, err := repo.SumDailyVolume()
	if err != nil {
		t.Fatalf("SumDailyVolume failed: %v", err)
	}
	if after < before+9.99 {
		t.Fatalf("expected daily volume to increase, before=%v after=%v", before, after)
	}

	reversalID := uuid.New().String()
	if err := repo.InsertReversal(reversalID, accountID, sql.NullString{String: toAccountID, Valid: true}, 5, txID); err != nil {
		t.Fatalf("InsertReversal failed: %v", err)
	}
	if err := repo.MarkTransactionReversed(txID); err != nil {
		t.Fatalf("MarkTransactionReversed failed: %v", err)
	}
	_, _ = db.Exec("DELETE FROM transactions WHERE transaction_id = $1", reversalID)
}

func openTestDB(t *testing.T) *sql.DB {
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "user")
	dbPass := getEnv("DB_PASS", "password")
	dbName := getEnv("DB_NAME", "banking")

	db, err := NewPostgresDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		t.Fatalf("NewPostgresDB failed: %v", err)
	}
	return db
}

func loadEnvForTests() {
	_ = config.LoadDotEnv(".env")
	_ = config.LoadDotEnv("..\\.env")
	_ = config.LoadDotEnv("..\\..\\.env")
	_ = config.LoadDotEnv("..\\..\\..\\.env")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
