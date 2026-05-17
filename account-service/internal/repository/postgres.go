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

type Account struct {
	AccountID string
	UserID    string
	Currency  string
	Balance   float64
	Status    string
	Tier      string
}

type Card struct {
	CardID    string
	AccountID string
	CardType  string
	Number    string
	Status    string
	Limit     float64
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

func (r *AccountRepository) ListAccountsByUser(userID string) ([]Account, error) {
	query := `SELECT account_id, user_id, currency, balance, status, tier FROM accounts WHERE user_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.AccountID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Status, &acc.Tier); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, rows.Err()
}

func (r *AccountRepository) UpdateBalance(accountID string, amount float64) error {
	query := `UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`
	_, err := r.db.Exec(query, amount, accountID)
	return err
}

func (r *AccountRepository) UpdateAccountStatus(accountID, status string) error {
	query := `UPDATE accounts SET status = $1 WHERE account_id = $2`
	_, err := r.db.Exec(query, status, accountID)
	return err
}

func (r *AccountRepository) UpdateAccountTier(accountID, tier string) error {
	query := `UPDATE accounts SET tier = $1 WHERE account_id = $2`
	_, err := r.db.Exec(query, tier, accountID)
	return err
}

func (r *AccountRepository) GetAccountLimits(accountID string) (float64, float64, error) {
	query := `SELECT daily_limit, monthly_limit FROM accounts WHERE account_id = $1`

	var daily, monthly float64
	if err := r.db.QueryRow(query, accountID).Scan(&daily, &monthly); err != nil {
		return 0, 0, err
	}
	return daily, monthly, nil
}

func (r *AccountRepository) CreateCard(cardID, accountID, cardType, number string) error {
	query := `INSERT INTO cards (card_id, account_id, card_type, number, status, limit_amount) 
			  VALUES ($1, $2, $3, $4, 'ACTIVE', 0.0)`
	_, err := r.db.Exec(query, cardID, accountID, cardType, number)
	return err
}

func (r *AccountRepository) GetCardByID(cardID string) (Card, error) {
	query := `SELECT card_id, account_id, card_type, number, status, limit_amount FROM cards WHERE card_id = $1`

	var card Card
	if err := r.db.QueryRow(query, cardID).Scan(&card.CardID, &card.AccountID, &card.CardType, &card.Number, &card.Status, &card.Limit); err != nil {
		return Card{}, err
	}
	return card, nil
}

func (r *AccountRepository) BlockCard(cardID string) error {
	query := `UPDATE cards SET status = 'BLOCKED' WHERE card_id = $1`
	_, err := r.db.Exec(query, cardID)
	return err
}

func (r *AccountRepository) SetCardLimit(cardID string, limit float64) error {
	query := `UPDATE cards SET limit_amount = $1 WHERE card_id = $2`
	_, err := r.db.Exec(query, limit, cardID)
	return err
}
