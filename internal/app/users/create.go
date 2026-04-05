package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	Users Creator
}

func (c *Create) Create(ctx context.Context, opts CreateOpts) (uuid.UUID, error) {
	if err := validateUsername(opts.Username); err != nil {
		return uuid.Nil, err
	}
	if err := validateLocale(opts.Locale); err != nil {
		return uuid.Nil, err
	}
	if err := validatePassword(opts.Password); err != nil {
		return uuid.Nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(opts.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	id, err := c.Users.Create(ctx, CreateDBOpts{
		Username:     opts.Username,
		Locale:       opts.Locale,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository: %w", err)
	}
	return id, nil
}
