package clients

import (
	"context"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

// Repo is the full client storage contract needed by mutating use cases.
// It is satisfied by *inframemory.ClientRepoCached (and the underlying postgres repo).
type Repo interface {
	Save(ctx context.Context, c *goidc.Client) error
	Client(ctx context.Context, id string) (*goidc.Client, error)
	Delete(ctx context.Context, id string) error
}

// Getter is the read-only subset needed by [GetByID].
type Getter interface {
	Client(ctx context.Context, id string) (*goidc.Client, error)
}
