package clients

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type Delete struct {
	Clients Repo
	Logger  *slog.Logger
}

func NewDelete(clients Repo, logger *slog.Logger) *Delete {
	if logger == nil {
		logger = slog.Default()
	}
	return &Delete{Clients: clients, Logger: logger}
}

// Delete removes the client with the given id.
// Returns an error marked with [errs.ErrNotFound] when no client matches,
// since the underlying SQL DELETE does not distinguish missing rows.
func (d *Delete) Delete(ctx context.Context, id string) error {
	if _, err := d.Clients.Client(ctx, id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.Rewrap("client not found", err)
		}
		d.Logger.ErrorContext(ctx, "delete client: check existence error", "id", id, "error", err)
		return fmt.Errorf("check existence: %w", err)
	}
	if err := d.Clients.Delete(ctx, id); err != nil {
		d.Logger.ErrorContext(ctx, "delete client: repository error", "id", id, "error", err)
		return fmt.Errorf("repository: %w", err)
	}
	return nil
}
