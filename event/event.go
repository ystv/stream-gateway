package event

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/stream-gateway/ticket"
	"github.com/ystv/stream-gateway/utils"
)

type (
	EventRepo interface {
		New(ctx context.Context, e NewEvent) (int, error)
	}
	Event struct {
		EventID string
		Name    string
		Tickets []ticket.Ticket
	}
	NewEvent struct {
		Name string
	}
	Eventer struct {
		db *sqlx.DB
	}
)

var _ EventRepo = &Eventer{}

// New creates an event
func (e *Eventer) New(ctx context.Context, event NewEvent) (int, error) {
	ticketID := 0
	err := e.db.GetContext(ctx, &ticketID, `
		INSERT INTO gateway.events(name)
		VALUES ($1) RETURNING event_id;`, event.Name)
	if err != nil {
		return ticketID, fmt.Errorf("failed to insert event: %w", err)
	}
	return ticketID, nil
}

// Delete a event
func (e *Eventer) Delete(ctx context.Context, eventID string) error {
	err := utils.Transact(e.db, func(tx *sqlx.Tx) error {
		res, err := tx.ExecContext(ctx, `
			DELETE FROM gateway.tickets
			WHERE event_id = $1`, eventID)
		if err != nil {
			return fmt.Errorf("failed to delete event \"%s\" tickets: %w", eventID, err)
		}
		changed, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to find rows affected: %w", err)
		}
		if changed == 0 {
			return errors.New("event doesn't exist")
		}

		res, err = tx.ExecContext(ctx, `
			DELETE FROM gateway.events
			WHERE event_id = $1`, eventID)
		if err != nil {
			return fmt.Errorf("failed to delete event info \"%s\": %w", eventID, err)
		}
		changed, err = res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to find rows affected: %w", err)
		}
		if changed == 0 {
			return errors.New("event doesn't exist")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}
