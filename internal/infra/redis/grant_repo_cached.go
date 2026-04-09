package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/redis/go-redis/v9"
)

// GrantTTL is the maximum time a grant is kept in the Redis cache.
// The actual TTL for a given grant is min(GrantTTL, time until expiry).
const GrantTTL = 60 * time.Second

const (
	grantKeyPrefix      = "grant:"
	grantAuthCodePrefix = "grant:authcode:"
)

// GrantRepoCached is a short-TTL Redis cache that wraps a [goidc.GrantManager].
//
// Only Grant-by-ID lookups are cached: they are the hot path during token
// introspection and userinfo calls. GrantByRefreshToken is less frequent and
// always goes straight to the underlying store.
//
// Key scheme:
//   - grant:{id}            → JSON-encoded *goidc.Grant, TTL = GrantTTL
//   - grant:authcode:{code} → grant ID (string), TTL = GrantTTL
//
// The secondary authcode key is written on Save (when AuthCode is set) so
// that DeleteByAuthCode can evict the primary key without an extra round-trip
// to the database.
type GrantRepoCached struct {
	repo   goidc.GrantManager
	client *redis.Client
}

// NewGrantRepoCached wraps repo with a Redis-backed cache layer.
func NewGrantRepoCached(repo goidc.GrantManager, client *redis.Client) *GrantRepoCached {
	return &GrantRepoCached{repo: repo, client: client}
}

// Save persists the grant to the underlying store and refreshes the cache entry.
// A failed cache write is silently ignored: the source of truth is the underlying store.
func (r *GrantRepoCached) Save(ctx context.Context, grant *goidc.Grant) error {
	if err := r.repo.Save(ctx, grant); err != nil {
		return err
	}

	data, err := json.Marshal(grant)
	if err != nil {
		return nil // should never happen; don't fail a successful Save
	}

	ttl := effectiveTTL(grant)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, grantKey(grant.ID), data, ttl)
	if grant.AuthCode != "" {
		pipe.Set(ctx, grantAuthCodeKey(grant.AuthCode), grant.ID, ttl)
	}
	pipe.Exec(ctx) //nolint:errcheck // cache write failure is non-fatal
	return nil
}

// Grant returns the grant with the given id, serving from cache when available.
// On a cache miss the grant is fetched from the underlying store and written back.
// Redis errors other than key-not-found are treated as cache misses.
func (r *GrantRepoCached) Grant(ctx context.Context, id string) (*goidc.Grant, error) {
	data, err := r.client.Get(ctx, grantKey(id)).Bytes()
	if err == nil {
		var grant goidc.Grant
		if err := json.Unmarshal(data, &grant); err == nil {
			return &grant, nil
		}
	}

	grant, err := r.repo.Grant(ctx, id)
	if err != nil {
		return nil, err
	}

	// write back; ignore cache errors
	if data, err := json.Marshal(grant); err == nil {
		r.client.Set(ctx, grantKey(id), data, effectiveTTL(grant)) //nolint:errcheck
	}
	return grant, nil
}

// GrantByRefreshToken always hits the underlying store directly.
// Refresh token exchanges are infrequent; caching them adds complexity for little gain.
func (r *GrantRepoCached) GrantByRefreshToken(ctx context.Context, token string) (*goidc.Grant, error) {
	return r.repo.GrantByRefreshToken(ctx, token)
}

// Delete removes the grant from the underlying store and evicts the cache entry.
func (r *GrantRepoCached) Delete(ctx context.Context, id string) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}
	r.client.Del(ctx, grantKey(id)) //nolint:errcheck
	return nil
}

// DeleteByAuthCode removes the grant associated with the given authorisation code
// from the underlying store. If the authcode-to-id mapping is present in Redis
// the corresponding primary cache entry is also evicted.
func (r *GrantRepoCached) DeleteByAuthCode(ctx context.Context, code string) error {
	// Look up the grant ID before deleting so we can evict the primary key.
	grantID, lookupErr := r.client.Get(ctx, grantAuthCodeKey(code)).Result()

	if err := r.repo.DeleteByAuthCode(ctx, code); err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, grantAuthCodeKey(code))
	if lookupErr == nil && grantID != "" {
		pipe.Del(ctx, grantKey(grantID))
	}
	pipe.Exec(ctx) //nolint:errcheck
	return nil
}

// grantKey returns the Redis key for a grant by id.
func grantKey(id string) string {
	return grantKeyPrefix + id
}

// grantAuthCodeKey returns the Redis key for the authcode→id secondary index.
func grantAuthCodeKey(code string) string {
	return grantAuthCodePrefix + code
}

// effectiveTTL returns the TTL to apply to a cache entry.
// It is the lesser of GrantTTL and the grant's remaining lifetime.
// Grants that never expire (ExpiresAtTimestamp == 0) always use GrantTTL.
func effectiveTTL(grant *goidc.Grant) time.Duration {
	if grant.ExpiresAtTimestamp == 0 {
		return GrantTTL
	}
	remaining := time.Until(time.Unix(int64(grant.ExpiresAtTimestamp), 0))
	if remaining <= 0 || remaining > GrantTTL {
		return GrantTTL
	}
	return remaining
}

var _ goidc.GrantManager = (*GrantRepoCached)(nil)
