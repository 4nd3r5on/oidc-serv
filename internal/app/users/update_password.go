package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type UpdatePasswordOpts struct {
	OldPassword string
	NewPassword string
}

func updatePassword(
	ctx context.Context,
	log *slog.Logger,
	users Updater,
	user *User,
	opts UpdatePasswordOpts,
) error {
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(opts.OldPassword)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		log.ErrorContext(ctx, "bcrypt compare failed", "error", err)
		return fmt.Errorf("compare password: %w", err)
	}

	if err := validatePassword(opts.NewPassword); err != nil {
		return err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(opts.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.ErrorContext(ctx, "bcrypt failed", "error", err)
		return fmt.Errorf("hash password: %w", err)
	}

	if err := users.Update(ctx, user.ID, UpdateDBOpts{Password: newHash}); err != nil {
		log.ErrorContext(ctx, "update password: repository error", "error", err)
		return fmt.Errorf("repository: %w", err)
	}
	return nil
}

type UpdatePasswordByIDRepo interface {
	GetterByID
	Updater
}

type UpdatePasswordByID struct {
	Users  UpdatePasswordByIDRepo
	Logger *slog.Logger
}

func NewUpdatePasswordByID(
	users UpdatePasswordByIDRepo,
	logger *slog.Logger,
) *UpdatePasswordByID {
	if logger == nil {
		logger = slog.Default()
	}
	return &UpdatePasswordByID{Users: users, Logger: logger}
}

func (up *UpdatePasswordByID) UpdatePasswordByID(
	ctx context.Context,
	id uuid.UUID,
	opts UpdatePasswordOpts,
) error {
	user, err := up.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.Rewrap("user not found", err)
		}
		up.Logger.ErrorContext(ctx, "update password: get user error", "error", err)
		return fmt.Errorf("repository: %w", err)
	}
	return updatePassword(ctx, up.Logger, up.Users, user, opts)
}
