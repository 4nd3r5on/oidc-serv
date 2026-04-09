package provider

import (
	"context"

	"github.com/4nd3r5on/oidc-serv/internal/app/users"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

// The following type aliases re-expose the goidc storage contracts so all
// infrastructure interfaces are visible from a single location.
// Implementations live in internal/infra/*.
type (
	ClientRepo        = goidc.ClientManager
	SessionRepo       = goidc.AuthnSessionManager
	GrantRepo         = goidc.GrantManager
	TokenRepo         = goidc.TokenManager
	LogoutSessionRepo = goidc.LogoutSessionManager
)

type UserPassMatcherFunc func(ctx context.Context, username, password string) (*users.User, error)
