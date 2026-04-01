// Package app provides application layer orchestration
package app

import (
	"context"
	"net/http"
	"strings"

	"github.com/4nd3r5on/oidc-serv/internal/app/users"
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"golang.org/x/crypto/bcrypt"
)

// StorageConfig groups the infrastructure implementations that back
// the OIDC provider with persistent storage.
// A nil field falls back to the provider's default in-memory storage.
type StorageConfig struct {
	Clients        ClientRepo
	Sessions       SessionRepo
	Grants         GrantRepo
	Tokens         TokenRepo
	LogoutSessions LogoutSessionRepo
}

// Service wires the OIDC provider with application domain logic.
type Service struct {
	*provider.Provider
	users users.Repository
}

// New creates a Service and applies storage backends and domain callbacks
// to the given provider.
func New(op *provider.Provider, userRepo users.Repository, store StorageConfig) (*Service, error) {
	svc := &Service{Provider: op, users: userRepo}
	return svc, op.WithOptions(svc.providerOpts(store)...)
}

func (s *Service) providerOpts(store StorageConfig) []provider.Option {
	opts := []provider.Option{
		provider.WithIDTokenClaims(s.idTokenClaims),
		provider.WithUserInfoClaims(s.userInfoClaims),
		provider.WithPolicies(s.authnPolicy()),
	}

	if store.Clients != nil {
		opts = append(opts, provider.WithClientManager(store.Clients))
	}
	if store.Sessions != nil {
		opts = append(opts, provider.WithAuthnSessionManager(store.Sessions))
	}
	if store.Grants != nil {
		opts = append(opts, provider.WithGrantManager(store.Grants))
	}
	if store.Tokens != nil {
		opts = append(opts, provider.WithTokenManager(store.Tokens))
	}
	if store.LogoutSessions != nil {
		opts = append(opts, provider.WithLogoutSessionManager(store.LogoutSessions))
	}

	return opts
}

func (s *Service) authnPolicy() goidc.AuthnPolicy {
	return goidc.NewPolicy(
		"password",
		func(_ *http.Request, _ *goidc.Client, _ *goidc.AuthnSession) bool { return true },
		s.authenticate,
	)
}

// authenticate validates username/password credentials submitted via form values.
func (s *Service) authenticate(_ http.ResponseWriter, r *http.Request, session *goidc.AuthnSession) (goidc.Status, error) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		return goidc.StatusFailure, nil
	}

	user, err := s.users.ByUsername(r.Context(), username)
	if err != nil {
		return goidc.StatusFailure, nil
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return goidc.StatusFailure, nil
	}

	session.SetUserID(user.ID.String())
	session.GrantScopes(session.Scopes)
	return goidc.StatusSuccess, nil
}

// idTokenClaims returns extra claims for the ID token.
func (s *Service) idTokenClaims(ctx context.Context, grant *goidc.Grant) map[string]any {
	return s.coreClaims(ctx, grant)
}

// userInfoClaims returns extra claims for the userinfo endpoint.
func (s *Service) userInfoClaims(ctx context.Context, grant *goidc.Grant) map[string]any {
	return s.coreClaims(ctx, grant)
}

// coreClaims returns username and locale claims when the "core" scope was granted.
func (s *Service) coreClaims(ctx context.Context, grant *goidc.Grant) map[string]any {
	if !strings.Contains(grant.Scopes, "core") {
		return nil
	}

	userID, err := uuid.Parse(grant.Subject)
	if err != nil {
		return nil
	}

	user, err := s.users.ByID(ctx, userID)
	if err != nil {
		return nil
	}

	return map[string]any{
		"username":        user.Username,
		goidc.ClaimLocale: user.Locale,
	}
}
