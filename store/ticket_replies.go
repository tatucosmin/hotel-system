package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TicketReplyStore struct {
	db *sqlx.DB
}

func NewTicketReplyStore(db *sql.DB) *TicketReplyStore {
	return &TicketReplyStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type TicketReply struct {
	Id        uuid.UUID `db:"id"`
	TicketId  uuid.UUID `db:"ticket_id"`
	Creator   uuid.UUID `db:"creator"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}

func (s *TicketReplyStore) Create(ctx context.Context, ticketId uuid.UUID, creatorId uuid.UUID, message string) (*TicketReply, error) {

	const query = `
	INSERT INTO ticket_replies (ticket_id, creator, message) VALUES ($1, $2, $3) RETURNING *`

	var ticketReply TicketReply
	if err := s.db.GetContext(ctx, &ticketReply, query, ticketId, creatorId, message); err != nil {
		return nil, fmt.Errorf("failed to create ticket reply: %w", err)
	}

	return &ticketReply, nil
}

func (s *TicketReplyStore) ByTicketId(ctx context.Context, ticketId uuid.UUID) (*[]TicketReply, error) {

	const query = `SELECT * FROM ticket_replies WHERE ticket_id = $1 ORDER BY created_at ASC`

	var ticketReplies []TicketReply
	if err := s.db.SelectContext(ctx, &ticketReplies, query, ticketId); err != nil {
		return nil, fmt.Errorf("failed to get ticket replies: %w", err)
	}

	return &ticketReplies, nil
}
