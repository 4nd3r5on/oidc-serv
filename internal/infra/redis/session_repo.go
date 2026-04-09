package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/redis/go-redis/v9"
)

// Key prefixes for authn sessions.
//
// Primary:   session:{id}                  → JSON AuthnSession
// Secondary: session:callback:{callbackID} → id
//            session:code:{authCode}        → id
//            session:par:{parID}            → id
//            session:ciba:{cibaID}          → id
//
// Secondary keys point to the primary so any lookup method can fetch the full
// session in two round-trips (GET secondary → GET primary). All keys carry the
// same TTL so they expire together.
const (
	sessionPrimaryPrefix  = "session:"
	sessionCallbackPrefix = "session:callback:"
	sessionCodePrefix     = "session:code:"
	sessionPARPrefix      = "session:par:"
	sessionCIBAPrefix     = "session:ciba:"
)

// SessionRepo is a Redis-only implementation of [goidc.AuthnSessionManager].
// Auth sessions are in-flight authorization flows (seconds to a few minutes);
// they require no durability beyond their TTL.
type SessionRepo struct {
	client *redis.Client
}

// NewSessionRepo returns a SessionRepo backed by the given Redis client.
func NewSessionRepo(client *redis.Client) *SessionRepo {
	return &SessionRepo{client: client}
}

// Save persists the session and all present secondary index keys.
// Not all secondaries exist at creation time — authCode is only added after
// the user authenticates — so Save may be called multiple times for the same
// session as it progresses through the flow.
func (r *SessionRepo) Save(ctx context.Context, session *goidc.AuthnSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := sessionTTL(session)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, sessionPrimaryKey(session.ID), data, ttl)
	if session.CallbackID != "" {
		pipe.Set(ctx, sessionCallbackKey(session.CallbackID), session.ID, ttl)
	}
	if session.AuthCode != "" {
		pipe.Set(ctx, sessionCodeKey(session.AuthCode), session.ID, ttl)
	}
	if session.PushedAuthReqID != "" {
		pipe.Set(ctx, sessionPARKey(session.PushedAuthReqID), session.ID, ttl)
	}
	if session.CIBAAuthID != "" {
		pipe.Set(ctx, sessionCIBAKey(session.CIBAAuthID), session.ID, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

// SessionByCallbackID fetches a session by its callback ID.
func (r *SessionRepo) SessionByCallbackID(ctx context.Context, callbackID string) (*goidc.AuthnSession, error) {
	return r.sessionBySecondaryKey(ctx, sessionCallbackKey(callbackID))
}

// SessionByAuthCode fetches a session by the authorisation code issued during
// the authorisation code flow.
func (r *SessionRepo) SessionByAuthCode(ctx context.Context, code string) (*goidc.AuthnSession, error) {
	return r.sessionBySecondaryKey(ctx, sessionCodeKey(code))
}

// SessionByPushedAuthReqID fetches a session by the PAR request URI id.
func (r *SessionRepo) SessionByPushedAuthReqID(ctx context.Context, id string) (*goidc.AuthnSession, error) {
	return r.sessionBySecondaryKey(ctx, sessionPARKey(id))
}

// SessionByCIBAAuthID fetches a session by the CIBA auth request ID.
func (r *SessionRepo) SessionByCIBAAuthID(ctx context.Context, id string) (*goidc.AuthnSession, error) {
	return r.sessionBySecondaryKey(ctx, sessionCIBAKey(id))
}

// Delete removes the primary session key and all secondary index keys.
// It fetches the session first to discover which secondaries are set.
func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	session, err := r.sessionByPrimaryKey(ctx, id)
	if err != nil {
		return err
	}

	keys := []string{sessionPrimaryKey(id)}
	if session.CallbackID != "" {
		keys = append(keys, sessionCallbackKey(session.CallbackID))
	}
	if session.AuthCode != "" {
		keys = append(keys, sessionCodeKey(session.AuthCode))
	}
	if session.PushedAuthReqID != "" {
		keys = append(keys, sessionPARKey(session.PushedAuthReqID))
	}
	if session.CIBAAuthID != "" {
		keys = append(keys, sessionCIBAKey(session.CIBAAuthID))
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// sessionBySecondaryKey resolves a secondary index key to an ID and then
// fetches the primary session entry.
func (r *SessionRepo) sessionBySecondaryKey(ctx context.Context, secondaryKey string) (*goidc.AuthnSession, error) {
	id, err := r.client.Get(ctx, secondaryKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("get session secondary key: %w", err)
	}
	return r.sessionByPrimaryKey(ctx, id)
}

// sessionByPrimaryKey fetches and deserialises the session stored under the
// primary key session:{id}.
func (r *SessionRepo) sessionByPrimaryKey(ctx context.Context, id string) (*goidc.AuthnSession, error) {
	data, err := r.client.Get(ctx, sessionPrimaryKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session goidc.AuthnSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &session, nil
}

func sessionPrimaryKey(id string) string         { return sessionPrimaryPrefix + id }
func sessionCallbackKey(callbackID string) string { return sessionCallbackPrefix + callbackID }
func sessionCodeKey(code string) string           { return sessionCodePrefix + code }
func sessionPARKey(parID string) string           { return sessionPARPrefix + parID }
func sessionCIBAKey(cibaID string) string         { return sessionCIBAPrefix + cibaID }

// sessionTTL returns the duration until the session expires.
// It floors at 1 second so Redis always gets a positive TTL.
func sessionTTL(session *goidc.AuthnSession) time.Duration {
	ttl := time.Until(time.Unix(int64(session.ExpiresAtTimestamp), 0))
	if ttl < time.Second {
		return time.Second
	}
	return ttl
}

var _ goidc.AuthnSessionManager = (*SessionRepo)(nil)
