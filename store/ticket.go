package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TicketPriority int

const (
	TicketPriorityUrgent TicketPriority = iota
	TicketPriorityHigh
	TicketPriorityMedium
	TicketPriorityLow
)

type TicketStore struct {
	db *sqlx.DB
}

type Ticket struct {
	Id              uuid.UUID      `db:"id"`
	Title           string         `db:"title"`
	Creator         uuid.UUID      `db:"creator"`
	CurrentAssignee uuid.UUID      `db:"current_assignee"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	Priority        TicketPriority `db:"priority"`
}

func NewTicketStore(db *sql.DB) *TicketStore {
	return &TicketStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *TicketStore) Create(ctx context.Context, title string, creatorId uuid.UUID, priority TicketPriority) (*Ticket, error) {

	const query = `
	INSERT INTO tickets (title, creator, priority) VALUES ($1, $2, $3) RETURNING *`

	var ticket Ticket
	if err := s.db.GetContext(ctx, &ticket, query, title, creatorId, priority); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return &ticket, nil
}

func (s *TicketStore) Delete(ctx context.Context, ticketId uuid.UUID) error {

	const query = `
	DELETE FROM tickets WHERE id = $1 RETURNING *`

	if _, err := s.db.ExecContext(ctx, query, ticketId); err != nil {
		return fmt.Errorf("failed to delete ticket with id %v: %w", ticketId, err)
	}

	return nil
}

func (s *TicketStore) Update(ctx context.Context, title string, ticketId, currentAssignee uuid.UUID, priority TicketPriority) error {

	const query = `
	UPDATE tickets SET title = $2, current_assignee = $3, updated_at = $4, priority = $5 WHERE id = $1`

	now := time.Now()

	if _, err := s.db.ExecContext(ctx, query, ticketId, title, currentAssignee, now, priority); err != nil {
		return fmt.Errorf("failed to update ticket with id %v: %w", ticketId, err)
	}

	return nil
}

func (s *TicketStore) ById(ctx context.Context, ticketId uuid.UUID) (*Ticket, error) {

	const query = `
	SELECT * FROM tickets WHERE id = $1`

	var ticket Ticket
	if err := s.db.GetContext(ctx, &ticket, query, ticketId); err != nil {
		return nil, fmt.Errorf("failed to get ticket with id %v: %w", ticketId, err)
	}

	return &ticket, nil
}

func (s *TicketStore) ByAssignee(ctx context.Context, currentAssignee uuid.UUID) (*Ticket, error) {

	const query = `
	SELECT * FROM tickets WHERE current_assignee = $1`

	var ticket Ticket
	if err := s.db.GetContext(ctx, &ticket, query, currentAssignee); err != nil {
		return nil, fmt.Errorf("failed to get tickets with assignee %v: %w", currentAssignee, err)
	}

	return &ticket, nil
}
