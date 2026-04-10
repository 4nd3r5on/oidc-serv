package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

var (
	ErrExpectedDataGotNil   = errors.New("expected auth data within client data, got nil")
	ErrMissingAuth          = errs.Mark(errors.New("authorization missing"), errs.ErrUnauthorized)
	ErrAuthenticationFailed = errs.Mark(errors.New("authentication failed"), errs.ErrUnauthorized)
)

type Method int

const (
	MethodNone Method = iota
	MethodUnknown
	MethodTMB // Trust me bro
	MethodSession
)

var methodStrMap = map[Method]string{
	MethodNone:    "",
	MethodUnknown: "unknown",
	MethodTMB:     "tmb",
	MethodSession: "session",
}

func (m Method) String() string {
	str, ok := methodStrMap[m]
	if !ok {
		return methodStrMap[MethodUnknown]
	}
	return str
}

// ClientData is the resolved authentication context attached to a request.
type ClientData struct {
	Method      Method
	TMBData     *TMBClaims
	SessionData *SessionClaims

	// Scopes usually contains static OpenAPI scopes
	Scopes []string
}

// Verifier verifies data existence/format validity and returns the client data
type Verifier interface {
	Verify(ctx context.Context, scheme, token string, scopes []string) (*ClientData, error)
}

// Core provides core for authentication based on the extracted client data
type Core interface {
	Auth(
		ctx context.Context,
		clientData *ClientData,
	) (userID uuid.UUID, authenticated bool, err error)
}

// None implements [Verifier] and [Core] for the [MethodNone]
type None struct{}

func (None) Verify(_ context.Context, _, _ string, scopes []string) (*ClientData, error) {
	return &ClientData{
		Method: MethodNone,
		Scopes: scopes,
	}, nil
}

func (None) Auth(
	_ context.Context,
	_ *ClientData,
) (userID uuid.UUID, authenticated bool, err error) {
	return uuid.Nil, false, nil
}

type Authenticator struct {
	Logger  *slog.Logger
	Methods map[Method]Core
}

func NewAuthenticator(logger *slog.Logger, methods map[Method]Core) *Authenticator {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("in", "Auth")
	if methods == nil {
		methods = make(map[Method]Core)
	}
	return &Authenticator{Logger: logger, Methods: methods}
}

func (a *Authenticator) callCore(ctx context.Context, clientData *ClientData, required bool) (uuid.UUID, error) {
	authCore, ok := a.Methods[clientData.Method]
	if !ok {
		err := fmt.Errorf("authentication core isn't implemented for the method %s",
			clientData.Method.String())
		err = errs.Mark(err, errs.ErrNotImplemented)
		a.Logger.ErrorContext(ctx, "no core registered for auth method",
			"method", clientData.Method.String())
		return uuid.Nil, err
	}
	userID, authenticated, err := authCore.Auth(ctx, clientData)
	if err != nil {
		a.Logger.WarnContext(ctx, "auth core returned error",
			"method", clientData.Method.String(), "error", err)
		return uuid.Nil, err
	}
	if !authenticated && required {
		return uuid.Nil, ErrAuthenticationFailed
	}
	return userID, nil
}

// handles nil client data then calls [Authenticator.callCore]
func (a *Authenticator) auth(ctx context.Context, required bool) (uuid.UUID, Method, error) {
	clientData, ok := CtxGetClientData(ctx)
	if !ok || clientData == nil {
		if !required {
			return uuid.Nil, MethodNone, nil
		}
		return uuid.Nil, MethodNone, ErrMissingAuth
	}
	userID, err := a.callCore(ctx, clientData, required)
	return userID, clientData.Method, err
}

// Auth implements [AuthFunc]
// wraps internal [Authenticator.auth] with additional error context
func (a *Authenticator) Auth(ctx context.Context, required bool) (uuid.UUID, error) {
	userID, method, err := a.auth(ctx, required)
	if err != nil {
		var errPrefix string
		if method != MethodNone {
			errPrefix = method.String() + " "
		}
		return uuid.Nil, fmt.Errorf("%sauth: %w", errPrefix, err)
	}
	return userID, nil
}
