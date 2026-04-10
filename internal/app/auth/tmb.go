package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type UserExister interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type TMBClaims struct {
	UserID uuid.UUID
}

// TMBVerifier provides [Verifier] function for TMB auth
type TMBVerifier struct {
	Logger *slog.Logger
}

func NewTMBVerifier(logger *slog.Logger) *TMBVerifier {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("in", "TMB Verifier")
	return &TMBVerifier{logger}
}

func (auth *TMBVerifier) Verify(ctx context.Context, _, token string, scopes []string) (*ClientData, error) {
	val, ok := strings.CutPrefix(token, "TMB ")
	if !ok {
		auth.Logger.DebugContext(ctx, "expected string with a TMB prefix", "got", token)
		return nil, errs.Mark(errs.ErrInvalidArgument)
	}
	userID, err := uuid.Parse(val)
	if err != nil {
		errMsg := "invalid user id format, expected UUID"
		auth.Logger.DebugContext(ctx, "TMB auth "+errMsg,
			"error", err, "value", val)
		return nil, errs.Mark(
			errors.New(errMsg),
			errs.ErrInvalidArgument,
		)
	}

	return &ClientData{
		Method: MethodTMB,
		TMBData: &TMBClaims{
			UserID: userID,
		},
		Scopes: scopes,
	}, nil
}

type TMBCore struct {
	Users UserExister
}

func (tmb *TMBCore) Auth(
	ctx context.Context,
	clientData *ClientData,
) (userID uuid.UUID, authenticated bool, err error) {
	if clientData.TMBData == nil {
		return uuid.Nil, false, ErrExpectedDataGotNil
	}
	exists, err := tmb.Users.Exists(ctx, clientData.TMBData.UserID)
	if err != nil {
		return uuid.Nil, false, err
	}
	if !exists {
		return uuid.Nil, false, nil
	}
	return clientData.TMBData.UserID, true, nil
}
