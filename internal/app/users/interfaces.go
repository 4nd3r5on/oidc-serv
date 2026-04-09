package users

import (
	"context"

	"github.com/google/uuid"
)

type Exister interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

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
	Password []byte // bcrypt hash, not plaintext
}

type Updater interface {
	Update(ctx context.Context, id uuid.UUID, opts UpdateDBOpts) error
}

type Deleter interface {
	Delete(ctx context.Context, id uuid.UUID) error
}
