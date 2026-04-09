package postgres

import (
	"context"

	"github.com/4nd3r5on/oidc-serv/pkg/db"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

// GrantRepo is a PostgreSQL-backed implementation of [goidc.GrantManager].
// The auth_details, resources, and store columns are stored as JSON blobs and
// marshalled/unmarshalled transparently by this repo.
type GrantRepo struct {
	q *db.Queries
}

// NewGrantRepo returns a GrantRepo backed by the given sqlc query set.
func NewGrantRepo(q *db.Queries) *GrantRepo {
	return &GrantRepo{q: q}
}

// Save upserts a grant, serialising variable-length fields to JSON columns.
func (r *GrantRepo) Save(ctx context.Context, g *goidc.Grant) error {
	authDetails, err := marshalJSON(g.AuthDetails)
	if err != nil {
		return err
	}
	resources, err := marshalJSON(g.Resources)
	if err != nil {
		return err
	}
	store, err := marshalJSON(g.Store)
	if err != nil {
		return err
	}
	return mapErr(r.q.UpsertGrant(ctx, db.UpsertGrantParams{
		ID:                   g.ID,
		RefreshToken:         toText(g.RefreshToken),
		AuthCode:             toText(g.AuthCode),
		GrantType:            string(g.Type),
		Subject:              g.Subject,
		ClientID:             g.ClientID,
		Scopes:               toText(g.Scopes),
		Nonce:                toText(g.Nonce),
		JwkThumbprint:        toText(g.JWKThumbprint),
		ClientCertThumbprint: toText(g.ClientCertThumbprint),
		AuthDetails:          authDetails,
		Resources:            resources,
		Store:                store,
		CreatedAt:            int64(g.CreatedAtTimestamp),
		ExpiresAt:            int64(g.ExpiresAtTimestamp),
	}))
}

// Grant fetches a grant by its primary key.
// Returns an error marked with [errs.ErrNotFound] when no row matches.
func (r *GrantRepo) Grant(ctx context.Context, id string) (*goidc.Grant, error) {
	row, err := r.q.GetGrant(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return rowToGrant(row)
}

// GrantByRefreshToken fetches the grant associated with the given refresh token.
// Returns an error marked with [errs.ErrNotFound] when no row matches.
func (r *GrantRepo) GrantByRefreshToken(ctx context.Context, token string) (*goidc.Grant, error) {
	row, err := r.q.GetGrantByRefreshToken(ctx, toText(token))
	if err != nil {
		return nil, mapErr(err)
	}
	return rowToGrant(row)
}

// Delete removes the grant row with the given id.
func (r *GrantRepo) Delete(ctx context.Context, id string) error {
	return mapErr(r.q.DeleteGrant(ctx, id))
}

// DeleteByAuthCode removes the grant associated with the given authorisation code.
func (r *GrantRepo) DeleteByAuthCode(ctx context.Context, code string) error {
	return mapErr(r.q.DeleteGrantByAuthCode(ctx, toText(code)))
}
