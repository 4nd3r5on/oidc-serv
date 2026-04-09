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
//	logout:{id}                  → JSON LogoutSession, TTL = ExpiresAtTimestamp - now
//	logout:callback:{callbackID} → id (string),        TTL = same
const (
	logoutPrimaryPrefix  = "logout:"
	logoutCallbackPrefix = "logout:callback:"
)

// LogoutSessionRepo is a Redis-only implementation of [goidc.LogoutSessionManager].
// Logout sessions are short-lived and require no durability beyond their TTL.
type LogoutSessionRepo struct {
	client *redis.Client
}

// NewLogoutSessionRepo returns a LogoutSessionRepo backed by the given Redis client.
func NewLogoutSessionRepo(client *redis.Client) *LogoutSessionRepo {
	return &LogoutSessionRepo{client: client}
}

// Save persists the logout session and its callback secondary index.
func (r *LogoutSessionRepo) Save(ctx context.Context, session *goidc.LogoutSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal logout session: %w", err)
	}

	ttl := logoutTTL(session)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, logoutPrimaryKey(session.ID), data, ttl)
	if session.CallbackID != "" {
		pipe.Set(ctx, logoutCallbackKey(session.CallbackID), session.ID, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("save logout session: %w", err)
	}
	return nil
}

// SessionByCallbackID fetches a logout session by its callback ID.
func (r *LogoutSessionRepo) SessionByCallbackID(ctx context.Context, callbackID string) (*goidc.LogoutSession, error) {
	id, err := r.client.Get(ctx, logoutCallbackKey(callbackID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("logout session not found")
		}
		return nil, fmt.Errorf("get logout session callback key: %w", err)
	}
	return r.sessionByID(ctx, id)
}

// Delete removes the logout session and its callback secondary index.
// It fetches the session first to discover the callbackID.
func (r *LogoutSessionRepo) Delete(ctx context.Context, id string) error {
	session, err := r.sessionByID(ctx, id)
	if err != nil {
		return err
	}

	keys := []string{logoutPrimaryKey(id)}
	if session.CallbackID != "" {
		keys = append(keys, logoutCallbackKey(session.CallbackID))
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("delete logout session: %w", err)
	}
	return nil
}

func (r *LogoutSessionRepo) sessionByID(ctx context.Context, id string) (*goidc.LogoutSession, error) {
	data, err := r.client.Get(ctx, logoutPrimaryKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("logout session not found")
		}
		return nil, fmt.Errorf("get logout session: %w", err)
	}

	var session goidc.LogoutSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal logout session: %w", err)
	}
	return &session, nil
}

func logoutPrimaryKey(id string) string           { return logoutPrimaryPrefix + id }
func logoutCallbackKey(callbackID string) string   { return logoutCallbackPrefix + callbackID }

func logoutTTL(session *goidc.LogoutSession) time.Duration {
	ttl := time.Until(time.Unix(int64(session.ExpiresAtTimestamp), 0))
	if ttl < time.Second {
		return time.Second
	}
	return ttl
}

var _ goidc.LogoutSessionManager = (*LogoutSessionRepo)(nil)
