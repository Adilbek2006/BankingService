package usecase

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"BankingService/user-service/internal/repository"
)

type fakeUserRepo struct {
	users      map[string]repository.User
	emailIndex map[string]string
	tokens     map[string]repository.UserToken
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:      map[string]repository.User{},
		emailIndex: map[string]string{},
		tokens:     map[string]repository.UserToken{},
	}
}

func (r *fakeUserRepo) SaveUser(userID, name, email, passwordHash string) error {
	r.users[userID] = repository.User{
		UserID:       userID,
		Name:         name,
		Email:        email,
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	r.emailIndex[email] = userID
	return nil
}

func (r *fakeUserRepo) GetUserByID(userID string) (repository.User, error) {
	u, ok := r.users[userID]
	if !ok {
		return repository.User{}, sql.ErrNoRows
	}
	return u, nil
}

func (r *fakeUserRepo) GetUserByEmail(email string) (repository.User, error) {
	id, ok := r.emailIndex[email]
	if !ok {
		return repository.User{}, sql.ErrNoRows
	}
	return r.users[id], nil
}

func (r *fakeUserRepo) UpdateUserName(userID, name string) error {
	u, ok := r.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	u.Name = name
	u.UpdatedAt = time.Now()
	r.users[userID] = u
	return nil
}

func (r *fakeUserRepo) UpdatePasswordHash(userID, passwordHash string) error {
	u, ok := r.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	u.PasswordHash = sql.NullString{String: passwordHash, Valid: true}
	u.UpdatedAt = time.Now()
	r.users[userID] = u
	return nil
}

func (r *fakeUserRepo) UpdateKYCStatus(userID, status string) error {
	u, ok := r.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	u.KYCStatus = sql.NullString{String: status, Valid: true}
	u.UpdatedAt = time.Now()
	r.users[userID] = u
	return nil
}

func (r *fakeUserRepo) SetSuspended(userID string, suspended bool) error {
	u, ok := r.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	u.IsSuspended = suspended
	u.UpdatedAt = time.Now()
	r.users[userID] = u
	return nil
}

func (r *fakeUserRepo) SetEmailVerified(userID string, verified bool) error {
	u, ok := r.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	u.EmailVerified = verified
	u.UpdatedAt = time.Now()
	r.users[userID] = u
	return nil
}

func (r *fakeUserRepo) DeleteUser(userID string) error {
	u, ok := r.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	delete(r.emailIndex, u.Email)
	delete(r.users, userID)
	return nil
}

func (r *fakeUserRepo) ListUsers() ([]repository.User, error) {
	users := make([]repository.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

func (r *fakeUserRepo) CreateToken(token, userID, tokenType string, expiresAt time.Time) error {
	r.tokens[token] = repository.UserToken{
		Token:     token,
		UserID:    userID,
		TokenType: tokenType,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}
	return nil
}

func (r *fakeUserRepo) GetValidToken(token, tokenType string) (repository.UserToken, error) {
	t, ok := r.tokens[token]
	if !ok {
		return repository.UserToken{}, sql.ErrNoRows
	}
	if t.TokenType != tokenType || t.Used || t.ExpiresAt.Before(time.Now()) {
		return repository.UserToken{}, sql.ErrNoRows
	}
	return t, nil
}

func (r *fakeUserRepo) MarkTokenUsed(token string) error {
	t, ok := r.tokens[token]
	if !ok {
		return sql.ErrNoRows
	}
	t.Used = true
	r.tokens[token] = t
	return nil
}

type fakeSender struct {
	lastTo      string
	lastSubject string
	lastBody    string
	err         error
}

func (s *fakeSender) Send(to, subject, body string) error {
	s.lastTo = to
	s.lastSubject = subject
	s.lastBody = body
	return s.err
}

func TestCreateUserSendsVerifyToken(t *testing.T) {
	repo := newFakeUserRepo()
	sender := &fakeSender{}
	uc := NewUserUsecase(repo, sender)

	userID, err := uc.CreateUser(context.Background(), "Ada", "ada@example.com", "Pass123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userID == "" {
		t.Fatal("expected userID")
	}
	if sender.lastTo != "ada@example.com" {
		t.Fatalf("unexpected email recipient: %s", sender.lastTo)
	}
	if !strings.Contains(sender.lastBody, "verification token") {
		t.Fatalf("expected verification token in email body")
	}

	found := false
	for _, tok := range repo.tokens {
		if tok.UserID == userID && tok.TokenType == tokenVerifyEmail {
			found = true
		}
	}
	if !found {
		t.Fatal("expected verification token stored")
	}
}

func TestChangePasswordInvalidOld(t *testing.T) {
	repo := newFakeUserRepo()
	sender := &fakeSender{}
	uc := NewUserUsecase(repo, sender)

	userID, err := uc.CreateUser(context.Background(), "Ada", "ada@example.com", "Pass123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	success, msg, err := uc.ChangePassword(context.Background(), userID, "wrong", "Pass456!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if success {
		t.Fatal("expected success=false")
	}
	if msg != "invalid_password" {
		t.Fatalf("unexpected msg: %s", msg)
	}
}

func TestVerifyEmail(t *testing.T) {
	repo := newFakeUserRepo()
	sender := &fakeSender{}
	uc := NewUserUsecase(repo, sender)

	userID, err := uc.CreateUser(context.Background(), "Ada", "ada@example.com", "Pass123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var token string
	for k, tok := range repo.tokens {
		if tok.UserID == userID && tok.TokenType == tokenVerifyEmail {
			token = k
			break
		}
	}
	if token == "" {
		t.Fatal("missing verify token")
	}

	success, msg, err := uc.VerifyEmail(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !success || msg != "email_verified" {
		t.Fatalf("unexpected result: %v %s", success, msg)
	}

	user, _ := repo.GetUserByID(userID)
	if !user.EmailVerified {
		t.Fatal("expected email verified")
	}
}

func TestResetPasswordFlow(t *testing.T) {
	repo := newFakeUserRepo()
	sender := &fakeSender{}
	uc := NewUserUsecase(repo, sender)

	_, err := uc.CreateUser(context.Background(), "Ada", "ada@example.com", "Pass123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, msg, err := uc.SendPasswordReset(context.Background(), "ada@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || msg != "reset_sent" {
		t.Fatalf("unexpected reset response: %v %s", ok, msg)
	}

	var resetToken string
	for k, tok := range repo.tokens {
		if tok.TokenType == tokenResetPassword {
			resetToken = k
			break
		}
	}
	if resetToken == "" {
		t.Fatal("missing reset token")
	}

	ok, msg, err = uc.ResetPassword(context.Background(), resetToken, "Pass456!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || msg != "password_reset" {
		t.Fatalf("unexpected reset result: %v %s", ok, msg)
	}
}

func TestSuspendUser(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUsecase(repo, &fakeSender{})

	userID, err := uc.CreateUser(context.Background(), "Ada", "ada@example.com", "Pass123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, msg, err := uc.SuspendUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || msg != "suspended" {
		t.Fatalf("unexpected suspend result: %v %s", ok, msg)
	}

	user, _ := repo.GetUserByID(userID)
	if !user.IsSuspended {
		t.Fatal("expected user suspended")
	}
}
