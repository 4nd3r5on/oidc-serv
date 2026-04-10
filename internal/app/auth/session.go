package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	appsession "github.com/4nd3r5on/oidc-serv/internal/app/session"
	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type SessionStore interface {
	Verify(ctx context.Context, sessionKey string) (*appsession.Data, error)
}

type SessionClaims struct {
	Key string
}

type VerifySession struct {
	Logger *slog.Logger
}

func NewVerifySession(logger *slog.Logger) *VerifySession {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("in", "Session Verifier")
	return &VerifySession{Logger: logger}
}

// Verify implements [Verifier].
// Does basic key validation and builds client data with the key.
func (v *VerifySession) Verify(ctx context.Context, _, token string, scopes []string) (*ClientData, error) {
	// NOTE: Might be later extended
	// for example, with TTL embedded into the token
	// for performance improvements
	// Logic: if the token says it's outdated we probably can just trust it?

	if err := appsession.ValidateKey(token); err != nil {
		v.Logger.DebugContext(ctx, "session key validation failed", "error", err)
		return nil, err
	}
	return &ClientData{
		Method: MethodSession,
		SessionData: &SessionClaims{
			Key: token,
		},
		Scopes: scopes,
	}, nil
}

type SessionCore struct {
	Verifier SessionStore
}

func (s *SessionCore) Auth(
	ctx context.Context,
	clientData *ClientData,
) (userID uuid.UUID, authenticated bool, err error) {
	if clientData.SessionData == nil {
		return uuid.Nil, false, ErrExpectedDataGotNil
	}
	data, err := s.Verifier.Verify(ctx, clientData.SessionData.Key)
	if err != nil {
		if errors.Is(err, errs.ErrUnauthorized) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return data.UserID, true, nil
}
