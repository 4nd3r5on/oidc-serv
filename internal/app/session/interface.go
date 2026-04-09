package session

import (
	"context"

	"github.com/google/uuid"
)

// Deleter removes a session by its opaque ID.
type Deleter interface {
	Delete(ctx context.Context, key string) error
}

// UserVerifyFunc authenticates a user by username and password,
// returning their ID on success.
type UserVerifyFunc func(ctx context.Context, username, password string) (uuid.UUID, error)

// Storer persists a session.
type Storer interface {
	Save(ctx context.Context, session Session) error
}

// Getter retrieves a session by its opaque ID.
type Getter interface {
	GetByKey(ctx context.Context, id string) (*Session, error)
}
