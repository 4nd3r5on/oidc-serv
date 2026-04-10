package api

import (
	"context"

	"github.com/google/uuid"

	appclients "github.com/4nd3r5on/oidc-serv/internal/app/clients"
	appusers "github.com/4nd3r5on/oidc-serv/internal/app/users"
)

type ClientCreator interface {
	Create(ctx context.Context, opts appclients.CreateOpts) (appclients.CreateRes, error)
}

type ClientGetterByID interface {
	Get(ctx context.Context, id string) (*appclients.Client, error)
}

type ClientDeleter interface {
	Delete(ctx context.Context, id string) error
}

type UserCreator interface {
	Create(ctx context.Context, opts appusers.CreateOpts) (uuid.UUID, error)
}

type UserGetterByID interface {
	Get(ctx context.Context, id uuid.UUID) (*appusers.GetRes, error)
}

type UserGetterByUsername interface {
	Get(ctx context.Context, username string) (*appusers.GetRes, error)
}

type MeService interface {
	Get(ctx context.Context) (*appusers.GetRes, error)
	Update(ctx context.Context, opts appusers.UpdateOpts) error
	Delete(ctx context.Context) error
	UpdatePassword(ctx context.Context, opts appusers.UpdatePasswordOpts) error
}
