package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tatucosmin/hotel-system/fixtures"
	"github.com/tatucosmin/hotel-system/store"
)

func TestUserStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	ctx := context.Background()

	cleanup := env.SetupDb(t)

	t.Cleanup(func() {
		cleanup(t)
	})

	userStore := store.NewUserStore(env.Db)

	user, err := userStore.CreateUser(ctx, "test@example.com", "testing")
	require.NoError(t, err)

	require.Equal(t, "test@example.com", user.Email)
	require.Equal(t, user.Roles, store.RoleCustomer)
	require.NoError(t, user.ComparePasswordHash("testing"))

	user.AddRole(store.RoleAdmin)
	user, err = userStore.UpdateUserById(ctx, user.Id, user.Email, user.Roles)
	require.NoError(t, err)
	require.True(t, user.HasRole(store.RoleAdmin))
	require.True(t, user.HasRole(store.RoleCustomer))
	require.False(t, user.HasRole(store.RoleStaff))

	user1, err := userStore.ById(ctx, user.Id)
	require.NoError(t, err)

	require.Equal(t, "test@example.com", user1.Email)
	require.Equal(t, user1.Roles, store.RoleCustomer|store.RoleAdmin)
	require.NoError(t, user1.ComparePasswordHash("testing"))
}
