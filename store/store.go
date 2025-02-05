package store

import "database/sql"

type Store struct {
	User         *UserStore
	RefreshToken *RefreshTokenStore
}

func New(db *sql.DB) *Store {
	return &Store{
		User:         NewUserStore(db),
		RefreshToken: NewRefreshTokenStore(db),
	}
}
