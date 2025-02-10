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

func (p TicketPriority) String() string {
	return []string{"urgent", "high", "medium", "low"}[p]
}

func (p TicketPriority) WithinBounds() bool {
	return p >= TicketPriorityUrgent && p <= TicketPriorityLow
}

type TicketStatus int

const (
	TicketStatusCreated TicketStatus = iota
	TicketStatusInProgress
	TicketStatusDone
	TicketStatusClosed
)

func (s TicketStatus) String() string {
	return []string{"created", "in_progress", "done", "closed"}[s]
}

func (s TicketStatus) WithinBounds() bool {
	return s >= TicketStatusCreated && s <= TicketStatusClosed
}

type TicketStore struct {
	db *sqlx.DB
}

type Ticket struct {
	Id              uuid.UUID      `db:"id"`
	Title           string         `db:"title"`
	Description     string         `db:"description"`
	Creator         uuid.UUID      `db:"creator"`
	CurrentAssignee uuid.UUID      `db:"current_assignee"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	Priority        TicketPriority `db:"priority"`
	Status          TicketStatus   `db:"status"`
}

func NewTicketStore(db *sql.DB) *TicketStore {
	return &TicketStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *TicketStore) Create(ctx context.Context, title, description string, creatorId uuid.UUID, priority TicketPriority) (*Ticket, error) {

	const query = `
	INSERT INTO tickets (title, description, creator, priority) VALUES ($1, $2, $3, $4) RETURNING *`

	var ticket Ticket
	if err := s.db.GetContext(ctx, &ticket, query, title, description, creatorId, priority); err != nil {
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

func (s *TicketStore) Update(ctx context.Context, ticketId uuid.UUID, priority TicketPriority, status TicketStatus) error {

	const query = `
	UPDATE tickets SET priority = $2, status = $3, updated_at = $4 WHERE id = $1`

	now := time.Now()

	if _, err := s.db.ExecContext(ctx, query, ticketId, priority, status, now); err != nil {
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

func (s *TicketStore) All(ctx context.Context) ([]Ticket, error) {

	const query = `
	SELECT * FROM tickets`

	var tickets []Ticket
	if err := s.db.SelectContext(ctx, &tickets, query); err != nil {
		return nil, fmt.Errorf("failed to get all tickets: %w", err)
	}

	return tickets, nil
}
