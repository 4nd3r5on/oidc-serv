package api

import (
	"context"
	"fmt"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/4nd3r5on/oidc-serv/internal/app/auth"
	"github.com/4nd3r5on/oidc-serv/pkg/api"
	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

type SecurityHandler struct {
	TMB, Session auth.Verifier
	AdminKey     string
}

func (sh *SecurityHandler) HandleSessionAuth(
	ctx context.Context,
	operationName api.OperationName,
	t api.SessionAuth,
) (context.Context, error) {
	if sh.Session == nil {
		return ctx, ogenerrors.ErrSkipServerSecurity
	}
	clientData, err := sh.Session.Verify(ctx, string(operationName), t.Token, t.Roles)
	return handleVerifierFuncOut(ctx, clientData, err)
}

func (sh *SecurityHandler) HandleAdminKeyAuth(
	ctx context.Context,
	_ api.OperationName,
	t api.AdminKeyAuth,
) (context.Context, error) {
	if sh.AdminKey == "" || t.APIKey != sh.AdminKey {
		return ctx, errs.ErrUnauthorized
	}
	return ctx, nil
}

func (sh *SecurityHandler) HandleTmbAuth(ctx context.Context, operationName api.OperationName, t api.TmbAuth) (context.Context, error) {
	if sh.TMB == nil {
		return ctx, ogenerrors.ErrSkipServerSecurity
	}
	clientData, err := sh.TMB.Verify(ctx, string(operationName), t.Token, t.Roles)
	return handleVerifierFuncOut(ctx, clientData, err)
}

func handleVerifierFuncOut(ctx context.Context, clientData *auth.ClientData, err error) (context.Context, error) {
	if err != nil {
		return ctx, authErr(err)
	}
	return auth.CtxPutClientData(ctx, clientData), nil
}

// authErr maps a verifier error to the appropriate ogen security error.
// Malformed tokens skip the current scheme so ogen can try the next one.
// Denied or internal errors fail the request immediately.
func authErr(err error) error {
	if errs.IsAny(err,
		errs.ErrInvalidArgument,
		errs.ErrMissingArgument,
		ogenerrors.ErrSkipServerSecurity,
	) {
		return ogenerrors.ErrSkipServerSecurity
	}
	if errs.IsAny(err, errs.ErrPermissionDenied, errs.ErrUnauthorized) {
		return err
	}
	return fmt.Errorf("auth: %w", err)
}
