package server

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tatucosmin/hotel-system/config"
)

type JwtManager struct {
	config *config.Config
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CClaims struct {
	Type string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewJwtManager(cfg *config.Config) *JwtManager {
	return &JwtManager{
		cfg,
	}
}

var SigningMethod = jwt.SigningMethodHS256

func (j *JwtManager) ParseToken(token string) (*jwt.Token, error) {
	parser := jwt.NewParser()
	jwtToken, err := parser.Parse(token, func(tk *jwt.Token) (any, error) {
		if tk.Method != SigningMethod {
			return nil, fmt.Errorf("expected signing method %v but got %v", SigningMethod.Alg(), tk.Method.Alg())
		}
		return []byte(j.config.JwtSecret), nil

	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return jwtToken, nil
}

func (j *JwtManager) IsAccessToken(token *jwt.Token) bool {
	jwtClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	if tokenType, ok := jwtClaims["token_type"]; ok {
		return tokenType == "access-token"
	}

	return false
}

func (j *JwtManager) GenerateTokens(userId uuid.UUID) (*TokenPair, error) {
	secret := []byte(j.config.JwtSecret)
	iss := "http://" + j.config.ServerHost + ":" + j.config.ServerPort
	now := time.Now()

	jwtAccessToken := jwt.NewWithClaims(SigningMethod, CClaims{
		Type: "access-token",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId.String(),
			Issuer:    iss,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 15)),
		},
	})

	signedAccessToken, err := jwtAccessToken.SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	accessToken, err := j.ParseToken(signedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	jwtRefreshToken := jwt.NewWithClaims(SigningMethod, CClaims{
		Type: "refresh-token",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId.String(),
			Issuer:    iss,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24 * 30)),
		},
	})

	signedRefreshToken, err := jwtRefreshToken.SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	refreshToken, err := j.ParseToken(signedRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}
