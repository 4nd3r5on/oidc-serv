package users

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type Exists struct {
	Logger *slog.Logger
	Users  Exister
}

func NewExists(users Exister, logger *slog.Logger) *Exists {
	if logger == nil {
		logger = slog.Default()
	}
	return &Exists{Users: users, Logger: logger}
}

func (e *Exists) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	exists, err := e.Users.Exists(ctx, id)
	if err != nil {
		e.Logger.ErrorContext(ctx, "user exists: repository error", "error", err)
		return false, fmt.Errorf("repository: %w", err)
	}
	return exists, nil
}
