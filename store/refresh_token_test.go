package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tatucosmin/hotel-system/fixtures"
	"github.com/tatucosmin/hotel-system/server"
	"github.com/tatucosmin/hotel-system/store"
)

func TestRefreshTokenStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	ctx := context.Background()

	cleanup := env.SetupDb(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	refreshTokenStore := store.NewRefreshTokenStore(env.Db)
	jwtManager := server.NewJwtManager(env.Config)
	userStore := store.NewUserStore(env.Db)

	user, err := userStore.CreateUser(ctx, "test@test.com", "test")
	require.NoError(t, err)

	tokens, err := jwtManager.GenerateTokens(user.Id)
	require.NoError(t, err)

	refreshToken, err := refreshTokenStore.Create(ctx, tokens.RefreshToken, user.Id)
	require.NoError(t, err)

	require.Equal(t, user.Id, refreshToken.UserId)
	require.NotEqual(t, tokens.RefreshToken.Raw, refreshToken.HashedToken)

	refreshToken, err = refreshTokenStore.ByPK(ctx, tokens.RefreshToken, user.Id)
	require.NoError(t, err)

	require.Equal(t, user.Id, refreshToken.UserId)
	require.NotEqual(t, tokens.RefreshToken.Raw, refreshToken.HashedToken)

	refreshTokenStore.RevokeAllFromUser(ctx, user.Id)

	refreshToken, err = refreshTokenStore.ByPK(ctx, tokens.RefreshToken, user.Id)
	require.Error(t, err)
	require.Nil(t, refreshToken)
}
