package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type UpdateOpts struct {
	Username *string
	Locale   *string
}

type Update struct {
	Users Updater
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
		return fmt.Errorf("repository: %w", err)
	}
	return nil
}
