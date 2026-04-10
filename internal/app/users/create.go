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

type CreateOpts struct {
	Username string
	Password string
	Locale   string
}

type CreateDBOpts struct {
	Username     string
	PasswordHash []byte
	Locale       string
}

type Create struct {
	Users  Creator
	Logger *slog.Logger
}

func NewCreate(users Creator, logger *slog.Logger) *Create {
	if logger == nil {
		logger = slog.Default()
	}
	return &Create{Users: users, Logger: logger}
}

func (c *Create) Create(ctx context.Context, opts CreateOpts) (uuid.UUID, error) {
	if err := validateUsername(opts.Username); err != nil {
		return uuid.Nil, err
	}
	if opts.Locale == "" {
		opts.Locale = "en"
	} else if err := validateLocale(opts.Locale); err != nil {
		return uuid.Nil, err
	}
	if err := validatePassword(opts.Password); err != nil {
		return uuid.Nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(opts.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger.ErrorContext(ctx, "bcrypt failed", "error", err)
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	id, err := c.Users.Create(ctx, CreateDBOpts{
		Username:     opts.Username,
		Locale:       opts.Locale,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, errs.ErrExists) {
			return uuid.Nil, errs.Rewrap("username already taken", err)
		}
		c.Logger.ErrorContext(ctx, "create user: repository error", "error", err)
		return uuid.Nil, fmt.Errorf("repository: %w", err)
	}
	return id, nil
}
