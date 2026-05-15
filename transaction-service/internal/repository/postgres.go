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
		return nil, err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS transactions (
		transaction_id VARCHAR(50) PRIMARY KEY,
		account_id VARCHAR(50) NOT NULL,
		to_account_id VARCHAR(50),
		amount DECIMAL(15, 2) NOT NULL,
		type VARCHAR(20) NOT NULL,
		status VARCHAR(20) DEFAULT 'SUCCESS',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactions table: %v", err)
	}

	log.Println("Transactions table ready")
	return db, nil
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) SaveTransaction(transactionID, accountID, txType string, amount float64) error {
	query := `INSERT INTO transactions (transaction_id, account_id, amount, type, status) VALUES ($1, $2, $3, $4, 'SUCCESS')`
	_, err := r.db.Exec(query, transactionID, accountID, amount, txType)
	return err
}

func (r *TransactionRepository) SaveTransfer(id, from, to string, amount float64) error {
	query := `INSERT INTO transactions (transaction_id, account_id, to_account_id, amount, type, status) 
              VALUES ($1, $2, $3, $4, 'TRANSFER', 'PENDING')`
	_, err := r.db.Exec(query, id, from, to, amount)
	return err
}
