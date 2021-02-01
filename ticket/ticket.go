package ticket

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type (
	TicketRepo interface {
		New(ctx context.Context, ticket NewTicket) (int, error)
		Get(ctx context.Context, token string) (Ticket, error)
		SetEnabled(ctx context.Context, ticketID string, enabled bool) error
		Delete(ctx context.Context, ticketID string) error
	}
	// Viewee represents a viewer
	Ticket struct {
		TicketID string
		Token    string
		Email    string
		Enabled  bool // isRevoked?
	}
	NewTicket struct {
		EventID string
		Email   string
		Enabled bool
	}
	Ticketer struct {
		db *sqlx.DB
	}
)

var _ TicketRepo = &Ticketer{}

func (t *Ticketer) New(ctx context.Context, ticket NewTicket) (int, error) {
	ticketID := 0
	err := t.db.GetContext(ctx, &ticketID, `
		INSERT INTO gateway.tickets(event_id, email, enabled)
		VALUES ($1, $2, $3) RETURNING ticket_id;`, ticket.EventID, ticket.Email, ticket.Enabled)
	if err != nil {
		return ticketID, fmt.Errorf("failed to insert ticket: %w", err)
	}
	return ticketID, nil
}

func (t *Ticketer) Get(ctx context.Context, token string) (Ticket, error) {
	ticket := Ticket{}
	err := t.db.GetContext(ctx, &ticket, `
		SELECT ticket_id, token, email, enabled
		FROM gateway.tickets
		WHERE token = $1;`, token)
	if err != nil {
		return ticket, fmt.Errorf("failed to get ticket: %w", err)
	}
	return ticket, nil
}

func (t *Ticketer) SetEnabled(ctx context.Context, ticketID string, enabled bool) error {
	res, err := t.db.ExecContext(ctx, `
		UPDATE SET
			enabled = $1
		WHERE ticket_id = $2;`, enabled, ticketID)
	if err != nil {
		return fmt.Errorf("failed to enable ticket \"%s\": %w", ticketID, err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to find rows affected: %w", err)
	}
	if changed == 0 {
		return errors.New("ticket doesn't exist")
	}
	return nil
}

func (t *Ticketer) Delete(ctx context.Context, ticketID string) error {
	res, err := t.db.ExecContext(ctx, `
		DELETE FROM gateway.tickets
		WHERE ticket_id = $1`, ticketID)
	if err != nil {
		return fmt.Errorf("failed to delete ticket \"%s\": %w", ticketID, err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to find rows affected: %w", err)
	}
	if changed == 0 {
		return errors.New("ticket doesn't exist")
	}
	return nil
}
