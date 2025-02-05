package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/tatucosmin/hotel-system/config"
)

func NewPgDatabase(cfg *config.Config) (*sql.DB, error) {
	connUrl := cfg.DatabaseUrl()
	db, err := sql.Open("postgres", connUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
