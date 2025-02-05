package fixtures

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/tatucosmin/hotel-system/config"
	"github.com/tatucosmin/hotel-system/store"
)

type TestEnv struct {
	Config *config.Config
	Db     *sql.DB
}

func NewTestEnv(t *testing.T) *TestEnv {
	os.Setenv("ENV", string(config.Env_Test))
	cfg, err := config.New()
	require.NoError(t, err)

	db, err := store.NewPgDatabase(cfg)
	require.NoError(t, err)

	return &TestEnv{
		Config: cfg,
		Db:     db,
	}
}

func (te *TestEnv) SetupDb(t *testing.T) func(t *testing.T) {
	err := goose.Up(te.Db, fmt.Sprintf("%s/%s", te.Config.ProjectRoot, "schema"))
	require.NoError(t, err)

	return te.CleanupDb
}

func (te *TestEnv) CleanupDb(t *testing.T) {
	err := goose.Down(te.Db, fmt.Sprintf("%s/%s", te.Config.ProjectRoot, "schema"))
	require.NoError(t, err)
}
