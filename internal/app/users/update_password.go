package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UpdatePasswordOpts struct {
	OldPassword string
	NewPassword string
}

type UpdatePasswordRepo interface {
	GetterByID
	Updater
}

type UpdatePassword struct {
	Users UpdatePasswordRepo
}

func (up *UpdatePassword) UpdatePassword(ctx context.Context, id uuid.UUID, opts UpdatePasswordOpts) error {
	user, err := up.Users.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("repository: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(opts.OldPassword)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("compare password: %w", err)
	}

	if err := validatePassword(opts.NewPassword); err != nil {
		return err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(opts.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := up.Users.Update(ctx, id, UpdateDBOpts{Password: newHash}); err != nil {
		return fmt.Errorf("repository: %w", err)
	}
	return nil
}
