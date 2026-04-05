package users

import (
	"context"

	"github.com/google/uuid"
)

type CreateUserOpts struct {
	Username string
	Password string
}

type CreateDBUserOpts struct {
	Username     string
	PasswordHash []byte
}

type Creator interface {
	Create(ctx context.Context, opts CreateDBUserOpts) (uuid.UUID, error)
}

type Create struct {
	Users Creator
}

func (c *Create) Create(ctx context.Context, opts CreateUserOpts) (uuid.UUID, error) {
	// TODO: Validate username
	id, err := c.Users.Create(ctx, CreateDBUserOpts{
		Username: opts.Username,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
