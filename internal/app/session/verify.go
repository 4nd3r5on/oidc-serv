package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type Verify struct {
	Sessions Getter
	Logger   *slog.Logger
}

func NewVerify(sessions Getter, logger *slog.Logger) *Verify {
	if logger == nil {
		logger = slog.Default()
	}
	return &Verify{Sessions: sessions, Logger: logger}
}

// Verify looks up the session by token, checks
// expiry, and returns ClientData populated with MethodSession.
// The scheme and scopes arguments are unused but kept for interface parity
// with other Verifier implementations.
func (v *Verify) Verify(ctx context.Context, sessionKey string) (*Data, error) {
	session, err := v.Sessions.GetByKey(ctx, sessionKey)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		v.Logger.ErrorContext(ctx, "verify session: get session error", "error", err)
		return nil, fmt.Errorf("get session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return &Data{
		UserID:     session.UserID,
		SessionKey: session.Key,
	}, nil
}
