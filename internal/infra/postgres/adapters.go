package postgres

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luikyv/go-oidc/pkg/goidc"

	"github.com/4nd3r5on/oidc-serv/internal/app/users"
	"github.com/4nd3r5on/oidc-serv/pkg/db"
)

// toText converts a plain string into a nullable [pgtype.Text].
// An empty string produces an invalid (NULL) value.
func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// rowToUser maps scalar columns returned by a sqlc user query to a domain [users.User].
func rowToUser(id pgtype.UUID, username string, passwordHash []byte, locale string) *users.User {
	return &users.User{
		ID:           uuid.UUID(id.Bytes),
		Username:     username,
		PasswordHash: passwordHash,
		Locale:       locale,
	}
}

// rowToGrant maps a [db.Grant] row to a domain [goidc.Grant],
// deserialising the JSON columns for auth details, resources, and store.
func rowToGrant(row db.Grant) (*goidc.Grant, error) {
	var authDetails []goidc.AuthDetail
	if err := unmarshalJSON(row.AuthDetails, &authDetails); err != nil {
		return nil, err
	}
	var resources goidc.Resources
	if err := unmarshalJSON(row.Resources, &resources); err != nil {
		return nil, err
	}
	var store map[string]any
	if err := unmarshalJSON(row.Store, &store); err != nil {
		return nil, err
	}
	return &goidc.Grant{
		ID:                   row.ID,
		RefreshToken:         row.RefreshToken.String,
		AuthCode:             row.AuthCode.String,
		Type:                 goidc.GrantType(row.GrantType),
		Subject:              row.Subject,
		ClientID:             row.ClientID,
		Scopes:               row.Scopes.String,
		Nonce:                row.Nonce.String,
		JWKThumbprint:        row.JwkThumbprint.String,
		ClientCertThumbprint: row.ClientCertThumbprint.String,
		AuthDetails:          authDetails,
		Resources:            resources,
		Store:                store,
		CreatedAtTimestamp:   int(row.CreatedAt),
		ExpiresAtTimestamp:   int(row.ExpiresAt),
	}, nil
}
