package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshTokenDb struct {
	UserId      uuid.UUID `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (r *RefreshTokenStore) generateBase64HashToken(token *jwt.Token) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(token.Raw))
	hashedBytes := hash.Sum(nil)

	encodedHashedToken := base64.StdEncoding.EncodeToString(hashedBytes)
	return encodedHashedToken, nil
}

func (r *RefreshTokenStore) Create(ctx context.Context, token *jwt.Token, userId uuid.UUID) (*RefreshTokenDb, error) {
	const query = `
	INSERT INTO refresh_tokens (user_id, hashed_token, expires_at) VALUES ($1, $2, $3) RETURNING *`

	encodedHashedToken, err := r.generateBase64HashToken(token)
	if err != nil {
		return nil, err
	}

	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get expiresAt from token: %w", err)
	}

	var refreshToken RefreshTokenDb
	if err := r.db.GetContext(ctx, &refreshToken, query, userId, encodedHashedToken, exp.Time); err != nil {
		return nil, fmt.Errorf("failed to create refreshToken inside database: %w", err)
	}

	return &refreshToken, nil
}

func (r *RefreshTokenStore) ByPK(ctx context.Context, token *jwt.Token, userId uuid.UUID) (*RefreshTokenDb, error) {
	const query = `
	SELECT * FROM refresh_tokens WHERE user_id = $1 AND hashed_token = $2`

	encodedHashedToken, err := r.generateBase64HashToken(token)
	if err != nil {
		return nil, err
	}

	var refreshToken RefreshTokenDb
	if err := r.db.GetContext(ctx, &refreshToken, query, userId, encodedHashedToken); err != nil {
		return nil, fmt.Errorf("failed to get refreshToken from PK: %w", err)
	}

	return &refreshToken, nil
}

func (r *RefreshTokenStore) RevokeAllFromUser(ctx context.Context, userId uuid.UUID) error {
	const query = `
	DELETE FROM refresh_tokens WHERE user_id = $1`

	if _, err := r.db.ExecContext(ctx, query, userId); err != nil {
		return fmt.Errorf("failed to revoke all tokens for %v: %w", userId, err)
	}

	return nil

}
