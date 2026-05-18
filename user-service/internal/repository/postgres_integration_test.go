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

func TestUserRepositoryCRUDIntegration(t *testing.T) {
	loadEnvForTests()
	cleanupTokensOnDelete(t)

	db := openTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New().String()
	email := "int-" + uuid.New().String() + "@example.com"

	if err := repo.SaveUser(userID, "Int User", email, "hash"); err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	user, err := repo.GetUserByID(userID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user.Email != email {
		t.Fatalf("unexpected email: %s", user.Email)
	}

	userByEmail, err := repo.GetUserByEmail(email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if userByEmail.UserID != userID {
		t.Fatalf("unexpected user id: %s", userByEmail.UserID)
	}

	if err := repo.UpdateUserName(userID, "New Name"); err != nil {
		t.Fatalf("UpdateUserName failed: %v", err)
	}
	if err := repo.UpdatePasswordHash(userID, "hash2"); err != nil {
		t.Fatalf("UpdatePasswordHash failed: %v", err)
	}
	if err := repo.UpdateKYCStatus(userID, "APPROVED"); err != nil {
		t.Fatalf("UpdateKYCStatus failed: %v", err)
	}
	if err := repo.SetSuspended(userID, true); err != nil {
		t.Fatalf("SetSuspended failed: %v", err)
	}
	if err := repo.SetEmailVerified(userID, true); err != nil {
		t.Fatalf("SetEmailVerified failed: %v", err)
	}

	if err := repo.DeleteUser(userID); err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

func TestUserRepositoryTokensIntegration(t *testing.T) {
	loadEnvForTests()
	cleanupTokensOnDelete(t)

	db := openTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New().String()
	email := "int-" + uuid.New().String() + "@example.com"

	if err := repo.SaveUser(userID, "Int User", email, "hash"); err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM user_tokens WHERE user_id = $1", userID)
		_ = repo.DeleteUser(userID)
	}()

	token := uuid.New().String()
	expires := time.Now().Add(10 * time.Minute)
	if err := repo.CreateToken(token, userID, "VERIFY_EMAIL", expires); err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	got, err := repo.GetValidToken(token, "VERIFY_EMAIL")
	if err != nil {
		t.Fatalf("GetValidToken failed: %v", err)
	}
	if got.UserID != userID {
		t.Fatalf("unexpected token user: %s", got.UserID)
	}

	if err := repo.MarkTokenUsed(token); err != nil {
		t.Fatalf("MarkTokenUsed failed: %v", err)
	}

	if _, err := repo.GetValidToken(token, "VERIFY_EMAIL"); err == nil {
		t.Fatal("expected invalid token after mark used")
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

func cleanupTokensOnDelete(t *testing.T) {
	t.Helper()
	// No-op: hooks left for clarity if future FK changes require setup.
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
