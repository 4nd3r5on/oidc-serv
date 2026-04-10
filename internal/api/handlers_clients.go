package api

import (
	"context"
	"errors"

	appclients "github.com/4nd3r5on/oidc-serv/internal/app/clients"
	genapi "github.com/4nd3r5on/oidc-serv/pkg/api"
	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

func (h *Handlers) CreateClient(ctx context.Context, req *genapi.CreateClientRequest) (genapi.CreateClientRes, error) {
	res, err := h.ClientCreate.Create(ctx, appclients.CreateOpts{
		ID:               req.ID,
		Secret:           req.Secret.Or(""),
		RedirectURIs:     urisToStrings(req.RedirectUris),
		GrantTypes:       req.GrantTypes,
		ResponseTypes:    req.ResponseTypes,
		ScopeIDs:         req.Scope.Or(""),
		TokenAuthnMethod: req.TokenEndpointAuthMethod.Or(""),
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidArgument) {
			return &genapi.CreateClientBadRequest{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrExists) {
			return conflictResponse(err), nil
		}
		return nil, err
	}
	return &genapi.CreateClientResponse{ID: res.ID, Secret: res.Secret}, nil
}

func (h *Handlers) GetClientById(ctx context.Context, params genapi.GetClientByIdParams) (genapi.GetClientByIdRes, error) { //nolint:revive
	c, err := h.ClientGet.Get(ctx, params.ClientId)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return &genapi.GetClientByIdNotFound{Error: err.Error()}, nil
		}
		return nil, err
	}
	return toOIDCClient(c), nil
}

func (h *Handlers) DeleteClient(ctx context.Context, params genapi.DeleteClientParams) (genapi.DeleteClientRes, error) {
	if err := h.ClientDelete.Delete(ctx, params.ClientId); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return &genapi.DeleteClientNotFound{Error: err.Error()}, nil
		}
		return nil, err
	}
	return &genapi.DeleteClientNoContent{}, nil
}
