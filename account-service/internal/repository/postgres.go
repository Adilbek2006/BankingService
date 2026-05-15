package repository

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgresDB(host, port, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	log.Println("PostgreSQL connection ready")
	return db, nil
}

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) SaveAccount(accountID, userID, currency string) error {
	query := `INSERT INTO accounts (account_id, user_id, currency, balance, status) 
			  VALUES ($1, $2, $3, 0.0, 'ACTIVE')`

	_, err := r.db.Exec(query, accountID, userID, currency)
	return err
}
func (r *AccountRepository) GetAccountByID(accountID string) (string, string, float64, string, error) {
	query := `SELECT user_id, currency, balance, status FROM accounts WHERE account_id = $1`

	var userID, currency, status string
	var balance float64

	err := r.db.QueryRow(query, accountID).Scan(&userID, &currency, &balance, &status)
	if err != nil {
		return "", "", 0, "", err
	}
	return userID, currency, balance, status, nil
}
func (r *AccountRepository) UpdateBalance(accountID string, amount float64) error {
	query := `UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`
	_, err := r.db.Exec(query, amount, accountID)
	return err
}
