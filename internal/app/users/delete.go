package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type Delete struct {
	Users  Deleter
	Logger *slog.Logger
}

func NewDelete(users Deleter, logger *slog.Logger) *Delete {
	if logger == nil {
		logger = slog.Default()
	}
	return &Delete{Users: users, Logger: logger}
}

func (d *Delete) Delete(ctx context.Context, id uuid.UUID) error {
	if err := d.Users.Delete(ctx, id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.Rewrap("user not found", err)
		}
		d.Logger.ErrorContext(ctx, "delete user: repository error", "error", err)
		return fmt.Errorf("repository: %w", err)
	}
	return nil
}
