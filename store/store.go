package store

import "database/sql"

type Store struct {
	User         *UserStore
	RefreshToken *RefreshTokenStore
	Ticket       *TicketStore
	TicketReply  *TicketReplyStore
}

func New(db *sql.DB) *Store {
	return &Store{
		User:         NewUserStore(db),
		RefreshToken: NewRefreshTokenStore(db),
		Ticket:       NewTicketStore(db),
		TicketReply:  NewTicketReplyStore(db),
	}
}
