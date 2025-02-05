package server_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tatucosmin/hotel-system/config"
	"github.com/tatucosmin/hotel-system/server"
)

func TestJwtManager(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	jwtManager := server.NewJwtManager(cfg)
	userId := uuid.New()

	tokens, err := jwtManager.GenerateTokens(userId)
	require.NoError(t, err)

	sbj, err := tokens.AccessToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), sbj)

	iss, err := tokens.AccessToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, "http://"+cfg.ServerHost+":"+cfg.ServerPort, iss)

	sbj, err = tokens.RefreshToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), sbj)

	iss, err = tokens.RefreshToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, "http://"+cfg.ServerHost+":"+cfg.ServerPort, iss)

	require.True(t, jwtManager.IsAccessToken(tokens.AccessToken))
	require.False(t, jwtManager.IsAccessToken(tokens.RefreshToken))

}
