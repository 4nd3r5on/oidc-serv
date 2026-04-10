package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type UpdateOpts struct {
	Username *string
	Locale   *string
}

type Update struct {
	Users  Updater
	Logger *slog.Logger
}

func NewUpdate(users Updater, logger *slog.Logger) *Update {
	if logger == nil {
		logger = slog.Default()
	}
	return &Update{Users: users, Logger: logger}
}

func (u *Update) Update(ctx context.Context, id uuid.UUID, opts UpdateOpts) error {
	if opts.Username != nil {
		if err := validateUsername(*opts.Username); err != nil {
			return err
		}
	}
	if opts.Locale != nil {
		if err := validateLocale(*opts.Locale); err != nil {
			return err
		}
	}
	if err := u.Users.Update(ctx, id, UpdateDBOpts{
		Username: opts.Username,
		Locale:   opts.Locale,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.Rewrap("user not found", err)
		}
		if errors.Is(err, errs.ErrExists) {
			return errs.Rewrap("username already taken", err)
		}
		u.Logger.ErrorContext(ctx, "update user: repository error", "error", err)
		return fmt.Errorf("repository: %w", err)
	}
	return nil
}
