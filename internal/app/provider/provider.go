// Package provider provides wrapping on top of [provider.Provider]
// providing app-specific functionality and helpers
package provider

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
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

type Provider struct {
	*provider.Provider
	users Users
}

// New creates a Provider and applies storage backends and domain callbacks
// to the given provider.
func New(
	op *provider.Provider,
	store StorageConfig,
	users Users,
) (*Provider, error) {
	svc := &Provider{Provider: op, users: users}
	return svc, op.WithOptions(svc.providerOpts(store)...)
}

func (p *Provider) providerOpts(store StorageConfig) []provider.Option {
	opts := []provider.Option{
		provider.WithIDTokenClaims(p.idTokenClaims),
		provider.WithUserInfoClaims(p.userInfoClaims),
		provider.WithPolicies(p.authnPolicy()),
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

func (p *Provider) authnPolicy() goidc.AuthnPolicy {
	return goidc.NewPolicy(
		"password",
		func(_ *http.Request, _ *goidc.Client, _ *goidc.AuthnSession) bool { return true },
		p.authenticate,
	)
}

// authenticate validates username/password credentials submitted via form values.
func (p *Provider) authenticate(_ http.ResponseWriter, r *http.Request, session *goidc.AuthnSession) (goidc.Status, error) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		return goidc.StatusFailure, nil
	}

	user, err := p.users.MatchUserPass(r.Context(), username, password)
	if err != nil {
		return goidc.StatusFailure, nil
	}

	session.SetUserID(user.ID.String())
	session.GrantScopes(session.Scopes)
	return goidc.StatusSuccess, nil
}

// idTokenClaims returns extra claims for the ID token.
func (p *Provider) idTokenClaims(ctx context.Context, grant *goidc.Grant) map[string]any {
	return p.coreClaims(ctx, grant)
}

// userInfoClaims returns extra claims for the userinfo endpoint.
func (p *Provider) userInfoClaims(ctx context.Context, grant *goidc.Grant) map[string]any {
	return p.coreClaims(ctx, grant)
}

// coreClaims returns username and locale claims when the "core" scope was granted.
func (p *Provider) coreClaims(ctx context.Context, grant *goidc.Grant) map[string]any {
	if !strings.Contains(grant.Scopes, "core") {
		return nil
	}

	userID, err := uuid.Parse(grant.Subject)
	if err != nil {
		return nil
	}

	user, err := p.users.ByID(ctx, userID)
	if err != nil {
		return nil
	}

	return map[string]any{
		"username":        user.Username,
		goidc.ClaimLocale: user.Locale,
	}
}
