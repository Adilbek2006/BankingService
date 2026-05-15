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
