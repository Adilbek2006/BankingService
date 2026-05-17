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

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(userID, name, email, passwordHash string) error {
	query := `INSERT INTO users (user_id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, userID, name, email, passwordHash)
	return err
}

type User struct {
	UserID        string
	Name          string
	Email         string
	PasswordHash  sql.NullString
	KYCStatus     sql.NullString
	IsSuspended   bool
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (r *UserRepository) GetUserByID(userID string) (User, error) {
	query := `SELECT user_id, name, email, password_hash, kyc_status, is_suspended, email_verified, created_at, updated_at
			  FROM users WHERE user_id = $1`
	var u User
	err := r.db.QueryRow(query, userID).Scan(
		&u.UserID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.KYCStatus,
		&u.IsSuspended,
		&u.EmailVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *UserRepository) GetUserByEmail(email string) (User, error) {
	query := `SELECT user_id, name, email, password_hash, kyc_status, is_suspended, email_verified, created_at, updated_at
			  FROM users WHERE email = $1`
	var u User
	err := r.db.QueryRow(query, email).Scan(
		&u.UserID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.KYCStatus,
		&u.IsSuspended,
		&u.EmailVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *UserRepository) UpdateUserName(userID, name string) error {
	query := `UPDATE users SET name = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.db.Exec(query, name, userID)
	return err
}

func (r *UserRepository) UpdatePasswordHash(userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.db.Exec(query, passwordHash, userID)
	return err
}

func (r *UserRepository) UpdateKYCStatus(userID, status string) error {
	query := `UPDATE users SET kyc_status = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.db.Exec(query, status, userID)
	return err
}

func (r *UserRepository) SetSuspended(userID string, suspended bool) error {
	query := `UPDATE users SET is_suspended = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.db.Exec(query, suspended, userID)
	return err
}

func (r *UserRepository) SetEmailVerified(userID string, verified bool) error {
	query := `UPDATE users SET email_verified = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.db.Exec(query, verified, userID)
	return err
}

func (r *UserRepository) DeleteUser(userID string) error {
	query := `DELETE FROM users WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *UserRepository) ListUsers() ([]User, error) {
	query := `SELECT user_id, name, email, password_hash, kyc_status, is_suspended, email_verified, created_at, updated_at
			  FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.UserID,
			&u.Name,
			&u.Email,
			&u.PasswordHash,
			&u.KYCStatus,
			&u.IsSuspended,
			&u.EmailVerified,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

type UserToken struct {
	Token     string
	UserID    string
	TokenType string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

func (r *UserRepository) CreateToken(token, userID, tokenType string, expiresAt time.Time) error {
	query := `INSERT INTO user_tokens (token, user_id, token_type, expires_at, used)
			  VALUES ($1, $2, $3, $4, false)`
	_, err := r.db.Exec(query, token, userID, tokenType, expiresAt)
	return err
}

func (r *UserRepository) GetValidToken(token, tokenType string) (UserToken, error) {
	query := `SELECT token, user_id, token_type, expires_at, used, created_at
			  FROM user_tokens
			  WHERE token = $1 AND token_type = $2 AND used = false AND expires_at > NOW()`
	var t UserToken
	err := r.db.QueryRow(query, token, tokenType).Scan(
		&t.Token,
		&t.UserID,
		&t.TokenType,
		&t.ExpiresAt,
		&t.Used,
		&t.CreatedAt,
	)
	if err != nil {
		return UserToken{}, err
	}
	return t, nil
}

func (r *UserRepository) MarkTokenUsed(token string) error {
	query := `UPDATE user_tokens SET used = true WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}
