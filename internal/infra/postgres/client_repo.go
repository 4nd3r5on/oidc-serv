package postgres

import (
	"context"
	"encoding/json"

	"github.com/luikyv/go-oidc/pkg/goidc"

	infracrypto "github.com/4nd3r5on/oidc-serv/internal/infra/crypto"
	"github.com/4nd3r5on/oidc-serv/pkg/db"
)

// clientRecord is the shape stored in the meta JSONB column.
// It captures all Client fields that are not stored in dedicated columns.
type clientRecord struct {
	FederationTrustAnchor string   `json:"federation_trust_anchor,omitempty"`
	FederationTrustMarks  []string `json:"federation_trust_marks,omitempty"`
	goidc.ClientMeta
}

// ClientRepo is a PostgreSQL-backed implementation of [goidc.ClientManager].
// The client secret and registration token are encrypted with AES-256-GCM
// before being persisted, and decrypted transparently on read.
type ClientRepo struct {
	q   *db.Queries
	key []byte // 32-byte AES-GCM key for Secret and RegistrationToken
}

// NewClientRepo returns a ClientRepo backed by the given sqlc query set.
// key must be exactly 32 bytes (AES-256).
func NewClientRepo(q *db.Queries, key []byte) *ClientRepo {
	return &ClientRepo{q: q, key: key}
}

// Save upserts a client, encrypting its secret and registration token before
// writing and serializing all remaining metadata to a JSONB column.
func (r *ClientRepo) Save(ctx context.Context, c *goidc.Client) error {
	secretEnc, err := infracrypto.Encrypt(r.key, []byte(c.Secret))
	if err != nil {
		return err
	}
	regTokenEnc, err := infracrypto.Encrypt(r.key, []byte(c.RegistrationToken))
	if err != nil {
		return err
	}
	meta, err := json.Marshal(clientRecord{
		FederationTrustAnchor: c.FederationTrustAnchor,
		FederationTrustMarks:  c.FederationTrustMarks,
		ClientMeta:            c.ClientMeta,
	})
	if err != nil {
		return err
	}
	return mapErr(r.q.UpsertClient(ctx, db.UpsertClientParams{
		ID:                   c.ID,
		SecretEnc:            secretEnc,
		RegistrationTokenEnc: regTokenEnc,
		CreatedAt:            int64(c.CreatedAtTimestamp),
		ExpiresAt:            int64(c.ExpiresAtTimestamp),
		Meta:                 meta,
	}))
}

// Client fetches and decrypts the client identified by id.
// Returns an error marked with [errs.ErrNotFound] when no row matches.
func (r *ClientRepo) Client(ctx context.Context, id string) (*goidc.Client, error) {
	row, err := r.q.GetClient(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}

	secretBytes, err := infracrypto.Decrypt(r.key, row.SecretEnc)
	if err != nil {
		return nil, err
	}
	regTokenBytes, err := infracrypto.Decrypt(r.key, row.RegistrationTokenEnc)
	if err != nil {
		return nil, err
	}

	var record clientRecord
	if err := json.Unmarshal(row.Meta, &record); err != nil {
		return nil, err
	}

	return &goidc.Client{
		ID:                    id,
		Secret:                string(secretBytes),
		RegistrationToken:     string(regTokenBytes),
		CreatedAtTimestamp:    int(row.CreatedAt),
		ExpiresAtTimestamp:    int(row.ExpiresAt),
		FederationTrustAnchor: record.FederationTrustAnchor,
		FederationTrustMarks:  record.FederationTrustMarks,
		ClientMeta:            record.ClientMeta,
	}, nil
}

// Delete removes the client row with the given id.
func (r *ClientRepo) Delete(ctx context.Context, id string) error {
	return mapErr(r.q.DeleteClient(ctx, id))
}
