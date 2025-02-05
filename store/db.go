package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tatucosmin/hotel-system/config"
)

func NewPgDatabase(cfg *config.Config) (*sqlx.DB, error) {
	connUrl := cfg.DatabaseUrl()
	db, err := sqlx.Open("postgres", connUrl)

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
