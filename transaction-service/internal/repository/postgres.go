package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

type Transaction struct {
	TransactionID string
	AccountID     string
	ToAccountID   sql.NullString
	Amount        float64
	Type          string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ReversedAt    sql.NullTime
	ReversalOf    sql.NullString
}

func (r *TransactionRepository) SaveTransaction(transactionID, accountID, txType string, amount float64, status string) error {
	query := `INSERT INTO transactions (transaction_id, account_id, amount, type, status)
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, transactionID, accountID, amount, txType, status)
	return err
}

func (r *TransactionRepository) SaveTransfer(id, from, to string, amount float64, status string) error {
	query := `INSERT INTO transactions (transaction_id, account_id, to_account_id, amount, type, status)
              VALUES ($1, $2, $3, $4, 'TRANSFER', $5)`
	_, err := r.db.Exec(query, id, from, to, amount, status)
	return err
}

func (r *TransactionRepository) GetTransactionByID(id string) (Transaction, error) {
	query := `SELECT transaction_id, account_id, to_account_id, amount, type, status, created_at, updated_at, reversed_at, reversal_of
              FROM transactions WHERE transaction_id = $1`

	var tx Transaction
	err := r.db.QueryRow(query, id).Scan(
		&tx.TransactionID,
		&tx.AccountID,
		&tx.ToAccountID,
		&tx.Amount,
		&tx.Type,
		&tx.Status,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&tx.ReversedAt,
		&tx.ReversalOf,
	)
	if err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func (r *TransactionRepository) UpdateTransactionStatus(id, status string) error {
	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE transaction_id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *TransactionRepository) MarkTransactionReversed(id string) error {
	query := `UPDATE transactions SET status = 'REVERSED', reversed_at = NOW(), updated_at = NOW() WHERE transaction_id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TransactionRepository) InsertReversal(reversalID, accountID string, toAccountID sql.NullString, amount float64, reversalOf string) error {
	query := `INSERT INTO transactions (transaction_id, account_id, to_account_id, amount, type, status, reversal_of)
              VALUES ($1, $2, $3, $4, 'REVERSAL', 'SUCCESS', $5)`
	_, err := r.db.Exec(query, reversalID, accountID, toAccountID, amount, reversalOf)
	return err
}

func (r *TransactionRepository) ListTransactionsByAccount(accountID string) ([]Transaction, error) {
	query := `SELECT transaction_id, account_id, to_account_id, amount, type, status, created_at, updated_at, reversed_at, reversal_of
              FROM transactions WHERE account_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.TransactionID,
			&tx.AccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Type,
			&tx.Status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&tx.ReversedAt,
			&tx.ReversalOf,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

func (r *TransactionRepository) ListPendingTransactions() ([]Transaction, error) {
	query := `SELECT transaction_id, account_id, to_account_id, amount, type, status, created_at, updated_at, reversed_at, reversal_of
              FROM transactions WHERE status = 'PENDING' ORDER BY created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.TransactionID,
			&tx.AccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Type,
			&tx.Status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&tx.ReversedAt,
			&tx.ReversalOf,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

func (r *TransactionRepository) ListTransactionsByAccountInRange(accountID string, start, end time.Time) ([]Transaction, error) {
	query := `SELECT transaction_id, account_id, to_account_id, amount, type, status, created_at, updated_at, reversed_at, reversal_of
              FROM transactions
              WHERE account_id = $1 AND created_at >= $2 AND created_at <= $3
              ORDER BY created_at ASC`

	rows, err := r.db.Query(query, accountID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.TransactionID,
			&tx.AccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Type,
			&tx.Status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&tx.ReversedAt,
			&tx.ReversalOf,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

func (r *TransactionRepository) SumDailyVolume() (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM transactions
              WHERE status = 'SUCCESS' AND created_at >= DATE_TRUNC('day', NOW())`

	var total float64
	if err := r.db.QueryRow(query).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
