package app

import (
	"github.com/4nd3r5on/oidc-serv/internal/app/users"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

// UserRepo is the storage interface for user accounts.
type UserRepo = users.Repository

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
