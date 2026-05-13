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
	CREATE TABLE IF NOT EXISTS accounts (
		account_id VARCHAR(50) PRIMARY KEY,
		user_id VARCHAR(50) NOT NULL,
		balance DECIMAL(15, 2) DEFAULT 0.0,
		currency VARCHAR(10) NOT NULL,
		status VARCHAR(20) DEFAULT 'ACTIVE',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("Error account table creation: %v", err)
	}

	log.Println(" Accounts table is ready to work")
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
