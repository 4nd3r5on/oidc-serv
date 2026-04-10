package clients

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type GetByID struct {
	Clients Getter
	Logger  *slog.Logger
}

func NewGetByID(clients Getter, logger *slog.Logger) *GetByID {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetByID{Clients: clients, Logger: logger}
}

func (g *GetByID) Get(ctx context.Context, id string) (*Client, error) {
	c, err := g.Clients.Client(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.Rewrap("client not found", err)
		}
		g.Logger.ErrorContext(ctx, "get client by id: repository error", "id", id, "error", err)
		return nil, fmt.Errorf("repository: %w", err)
	}
	return clientFromGoidc(c), nil
}
