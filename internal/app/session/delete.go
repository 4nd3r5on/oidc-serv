package session

import (
	"context"
	"fmt"
	"log/slog"
)

type Delete struct {
	Sessions Deleter
	Logger   *slog.Logger
}

func NewDelete(sessions Deleter, logger *slog.Logger) *Delete {
	if logger == nil {
		logger = slog.Default()
	}
	return &Delete{Sessions: sessions, Logger: logger}
}

func (uc *Delete) Delete(ctx context.Context, sessionID string) error {
	if err := uc.Sessions.Delete(ctx, sessionID); err != nil {
		uc.Logger.ErrorContext(ctx, "delete session: repository error", "error", err)
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
