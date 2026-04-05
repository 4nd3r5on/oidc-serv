package users

import (
	"context"

	"github.com/google/uuid"
)

type Creator interface {
	Create(ctx context.Context, opts CreateDBOpts) (uuid.UUID, error)
}

type GetterByID interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type GetterByUsername interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
}

type UpdateDBOpts struct {
	Username *string
	Locale   *string
	Password []byte
}

type Updater interface {
	Update(ctx context.Context, id uuid.UUID, opts UpdateDBOpts) error
}

type Deleter interface {
	Delete(ctx context.Context, id uuid.UUID) error
}
