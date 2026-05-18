//go:build integration
// +build integration

package repository

import (
	"database/sql"
	"os"
	"testing"

	"BankingService/internal/config"

	"github.com/google/uuid"
)

func TestAccountRepositoryIntegration(t *testing.T) {
	loadEnvForTests()

	db := openTestDB(t)
	defer db.Close()

	repo := NewAccountRepository(db)

	accountID := uuid.New().String()
	userID := uuid.New().String()

	if err := repo.SaveAccount(accountID, userID, "USD"); err != nil {
		t.Fatalf("SaveAccount failed: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM cards WHERE account_id = $1", accountID)
		_, _ = db.Exec("DELETE FROM accounts WHERE account_id = $1", accountID)
	}()

	_, _, _, status, err := repo.GetAccountByID(accountID)
	if err != nil {
		t.Fatalf("GetAccountByID failed: %v", err)
	}
	if status != "ACTIVE" {
		t.Fatalf("unexpected status: %s", status)
	}

	if err := repo.UpdateAccountStatus(accountID, "FROZEN"); err != nil {
		t.Fatalf("UpdateAccountStatus failed: %v", err)
	}
	if err := repo.UpdateAccountTier(accountID, "GOLD"); err != nil {
		t.Fatalf("UpdateAccountTier failed: %v", err)
	}

	_, _, err = repo.GetAccountLimits(accountID)
	if err != nil {
		t.Fatalf("GetAccountLimits failed: %v", err)
	}

	cardID := uuid.New().String()
	if err := repo.CreateCard(cardID, accountID, "VIRTUAL", "4111111111111111"); err != nil {
		t.Fatalf("CreateCard failed: %v", err)
	}
	if _, err := repo.GetCardByID(cardID); err != nil {
		t.Fatalf("GetCardByID failed: %v", err)
	}
	if err := repo.SetCardLimit(cardID, 1000); err != nil {
		t.Fatalf("SetCardLimit failed: %v", err)
	}
	if err := repo.BlockCard(cardID); err != nil {
		t.Fatalf("BlockCard failed: %v", err)
	}
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
