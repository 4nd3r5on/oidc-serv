package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4nd3r5on/oidc-serv/internal/app/session"
)

// Key scheme:
//
//	user_session:{id} → JSON userSessionRecord, TTL = ExpiresAt - now
const userSessionPrefix = "user_session:"

// userSessionRecord is the shape written to Redis.
// uuid.UUID does not serialize cleanly as JSON so we store it as a string.
type userSessionRecord struct {
	Key       string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UserSessionRepo is a Redis-backed implementation of the auth session
// repository interfaces ([auth.SessionStorer], [auth.SessionGetter],
// [auth.SessionDeleter]).
type UserSessionRepo struct {
	client *redis.Client
}

// NewUserSessionRepo returns a UserSessionRepo backed by the given Redis client.
func NewUserSessionRepo(client *redis.Client) *UserSessionRepo {
	return &UserSessionRepo{client: client}
}

// Save persists the session with a TTL derived from ExpiresAt.
func (r *UserSessionRepo) Save(ctx context.Context, s session.Session) error {
	rec := userSessionRecord{
		Key:       s.Key,
		UserID:    s.UserID.String(),
		ExpiresAt: s.ExpiresAt,
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal user session: %w", err)
	}

	ttl := max(time.Until(s.ExpiresAt), time.Second)

	if err := r.client.Set(ctx, userSessionKey(s.Key), data, ttl).Err(); err != nil {
		return fmt.Errorf("save user session: %w", err)
	}
	return nil
}

// GetByKey fetches the session with the given opaque token key.
func (r *UserSessionRepo) GetByKey(ctx context.Context, id string) (*session.Session, error) {
	data, err := r.client.Get(ctx, userSessionKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user session not found")
		}
		return nil, fmt.Errorf("get user session: %w", err)
	}

	var rec userSessionRecord
	if err = json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("unmarshal user session: %w", err)
	}

	userID, err := uuid.Parse(rec.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user session user_id: %w", err)
	}

	return &session.Session{
		Key:       rec.Key,
		UserID:    userID,
		ExpiresAt: rec.ExpiresAt,
	}, nil
}

// Delete removes the session with the given id.
func (r *UserSessionRepo) Delete(ctx context.Context, id string) error {
	if err := r.client.Del(ctx, userSessionKey(id)).Err(); err != nil {
		return fmt.Errorf("delete user session: %w", err)
	}
	return nil
}

func userSessionKey(id string) string { return userSessionPrefix + id }
