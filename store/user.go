package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type UserRole int

const (
	RoleAdmin    UserRole = 1 << iota // 1
	RoleStaff                         // 2
	RoleCustomer                      // 4
)

type User struct {
	Id             uuid.UUID `db:"id"`
	Email          string    `db:"email"`
	HashedPassword string    `db:"hashed_password"`
	CreatedAt      time.Time `db:"created_at"`
	Roles          UserRole  `db:"roles"`
}

func (u *User) ComparePasswordHash(password string) error {
	hashedPassword, err := base64.StdEncoding.DecodeString(u.HashedPassword)

	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		return fmt.Errorf("passwords do not match")
	}

	return nil
}

func (u *User) HasRole(role UserRole) bool {
	return u.Roles&role != 0
}

func (u *User) AddRole(role UserRole) {
	u.Roles |= role
}

func (u *User) RemoveRole(role UserRole) {
	u.Roles &^= role
}

func (s *UserStore) CreateUser(ctx context.Context, email, password string) (*User, error) {
	const query = `
	INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING *`

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("failed to generate password hash: %w", err)
	}

	hashedPassword := base64.StdEncoding.EncodeToString(bytes)

	var user User
	if err := s.db.GetContext(ctx, &user, query, email, hashedPassword); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)

	}

	return &user, nil
}

func (s *UserStore) UpdateUserById(ctx context.Context, userId uuid.UUID, email string, roles UserRole) (*User, error) {
	const query = `
	UPDATE users SET email = $1, roles = $2 WHERE id = $3 RETURNING *`

	var user User
	if err := s.db.GetContext(ctx, &user, query, email, roles, userId); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (s *UserStore) ById(ctx context.Context, userId uuid.UUID) (*User, error) {
	const query = `
	SELECT * FROM users WHERE id = $1`

	var user User
	if err := s.db.GetContext(ctx, &user, query, userId); err != nil {
		return nil, fmt.Errorf("failed to fetch user with id %s: %w", userId, err)
	}

	return &user, nil
}

func (s *UserStore) ByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
	SELECT * FROM users WHERE email = $1`

	var user User
	if err := s.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, fmt.Errorf("failed to fetch user with email %s: %w", email, err)
	}

	return &user, nil
}
