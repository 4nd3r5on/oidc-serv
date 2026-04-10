package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

const (
	SessionIDBytesLen = 32
	DefaultSessionTTL = 24 * time.Hour
)

type IssueSessionOpts struct {
	Username string
	Password string

	SessionTTL *time.Duration
}

type IssueSession struct {
	Users    UserVerifyFunc
	Sessions Storer
	Logger   *slog.Logger
}

func NewIssueSession(users UserVerifyFunc, sessions Storer, _ time.Duration, logger *slog.Logger) *IssueSession {
	if logger == nil {
		logger = slog.Default()
	}
	return &IssueSession{Users: users, Sessions: sessions, Logger: logger}
}

func (uc *IssueSession) IssueSession(ctx context.Context, opts IssueSessionOpts) (string, error) {
	userID, err := uc.Users(ctx, opts.Username, opts.Password)
	if err != nil {
		if errs.IsAny(err, errs.ErrUnauthorized, errs.ErrNotFound) {
			return "", err
		}
		uc.Logger.ErrorContext(ctx, "issue session: verify credentials error", "error", err)
		return "", fmt.Errorf("verify credentials: %w", err)
	}

	id, err := generateSessionID()
	if err != nil {
		uc.Logger.ErrorContext(ctx, "issue session: generate session id error", "error", err)
		return "", fmt.Errorf("generate session id: %w", err)
	}

	var ttl time.Duration
	if opts.SessionTTL == nil {
		ttl = DefaultSessionTTL
	} else {
		ttl = *opts.SessionTTL
	}

	if err := uc.Sessions.Save(ctx, Session{
		Key:       id,
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}); err != nil {
		uc.Logger.ErrorContext(ctx, "issue session: store session error", "error", err)
		return "", fmt.Errorf("store session: %w", err)
	}

	return id, nil
}

// generateSessionID returns a 64-char hex string backed by 32 random bytes.
func generateSessionID() (string, error) {
	b := make([]byte, SessionIDBytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
