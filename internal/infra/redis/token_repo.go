package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/redis/go-redis/v9"
)

// Key scheme:
//
//	token:{id}              → JSON Token,   TTL = ExpiresAtTimestamp - now
//	grant_tokens:{grantID}  → SET of token IDs, same TTL as the token
//
// The secondary set enables O(1) cascade revocation: DeleteByGrantID fetches
// all token IDs from the set and deletes them together with the set itself in
// a single pipeline.
const (
	tokenKeyPrefix      = "token:"
	grantTokensPrefix   = "grant_tokens:"
)

// TokenRepo is a Redis-backed implementation of [goidc.TokenManager].
// Tokens are stored with a TTL derived from their expiry timestamp so Redis
// handles expiration automatically without a cleanup job.
type TokenRepo struct {
	client *redis.Client
}

// NewTokenRepo returns a TokenRepo backed by the given Redis client.
func NewTokenRepo(client *redis.Client) *TokenRepo {
	return &TokenRepo{client: client}
}

// Save persists the token and registers its ID in the per-grant set so that
// DeleteByGrantID can revoke all tokens for a grant in one operation.
func (r *TokenRepo) Save(ctx context.Context, t *goidc.Token) error {
	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}

	ttl := tokenTTL(t)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, tokenKey(t.ID), data, ttl)
	pipe.SAdd(ctx, grantTokensKey(t.GrantID), t.ID)
	pipe.Expire(ctx, grantTokensKey(t.GrantID), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	return nil
}

// Token fetches the token with the given id.
func (r *TokenRepo) Token(ctx context.Context, id string) (*goidc.Token, error) {
	data, err := r.client.Get(ctx, tokenKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("get token: %w", err)
	}

	var t goidc.Token
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("unmarshal token: %w", err)
	}
	return &t, nil
}

// Delete removes a single token by id.
func (r *TokenRepo) Delete(ctx context.Context, id string) error {
	if err := r.client.Del(ctx, tokenKey(id)).Err(); err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}

// DeleteByGrantID revokes all tokens issued under the given grant.
// It fetches the grant's token-ID set, then deletes every token key and the
// set itself in a single pipeline.
func (r *TokenRepo) DeleteByGrantID(ctx context.Context, grantID string) error {
	ids, err := r.client.SMembers(ctx, grantTokensKey(grantID)).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("get grant tokens: %w", err)
	}

	keys := make([]string, 0, len(ids)+1)
	for _, id := range ids {
		keys = append(keys, tokenKey(id))
	}
	keys = append(keys, grantTokensKey(grantID))

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("delete grant tokens: %w", err)
	}
	return nil
}

func tokenKey(id string) string           { return tokenKeyPrefix + id }
func grantTokensKey(grantID string) string { return grantTokensPrefix + grantID }

// tokenTTL returns the duration until the token expires, floored at 1 second.
func tokenTTL(t *goidc.Token) time.Duration {
	ttl := time.Until(time.Unix(int64(t.ExpiresAtTimestamp), 0))
	if ttl < time.Second {
		return time.Second
	}
	return ttl
}

var _ goidc.TokenManager = (*TokenRepo)(nil)
