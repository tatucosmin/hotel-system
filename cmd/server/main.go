package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tatucosmin/hotel-system/config"
	"github.com/tatucosmin/hotel-system/server"
	"github.com/tatucosmin/hotel-system/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		return err
	}

	db, err := store.NewPgDatabase(cfg)
	if err != nil {
		return err
	}
	store := store.New(db)

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)

	jwtManager := server.NewJwtManager(cfg)

	server := server.New(cfg, logger, store, jwtManager)
	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil

}
