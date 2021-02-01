package watch

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/stream-gateway/ticket"
)

type (
	Watcher struct {
		db       *sqlx.DB
		tck      *ticket.Ticketer
		watchers map[string]*Watch
	}
	Watch struct {
		RemoteAddr string
		EventID    string
		Ticket     ticket.Ticket
	}
)

func (w *Watcher) NewWatch(ctx context.Context, r *http.Request, token string) (*Watch, error) {
	watch, ok := w.watchers[token]
	if ok {
		// Only one ticket can be used at a time, so error
		return watch, errors.New("ticket in-use")
	}
	eventID := ""
	t, err := w.tck.Get(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	// Not watching yet, add these to the watchers map
	watch = &Watch{
		RemoteAddr: r.RemoteAddr,
		EventID:    eventID,
		Ticket:     t,
	}
	w.watchers[token] = watch
	return watch, nil
}
