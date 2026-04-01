package users

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines storage operations for user accounts.
type Repository interface {
	ByID(ctx context.Context, id uuid.UUID) (*User, error)
	ByUsername(ctx context.Context, username string) (*User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ServiceInterface interface {
	Create(ctx context.Context) error
	Update(ctx context.Context) error
}
