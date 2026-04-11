package main

import (
	"context"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/4nd3r5on/oidc-serv/pkg/api"
)

type adminSecurity struct{ key string }

func (s *adminSecurity) AdminKeyAuth(_ context.Context, _ api.OperationName) (api.AdminKeyAuth, error) {
	return api.AdminKeyAuth{APIKey: s.key}, nil
}

func (s *adminSecurity) SessionAuth(_ context.Context, _ api.OperationName) (api.SessionAuth, error) {
	return api.SessionAuth{}, ogenerrors.ErrSkipClientSecurity
}

func (s *adminSecurity) TmbAuth(_ context.Context, _ api.OperationName) (api.TmbAuth, error) {
	return api.TmbAuth{}, ogenerrors.ErrSkipClientSecurity
}
