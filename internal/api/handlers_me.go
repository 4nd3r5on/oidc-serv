package api

import (
	"context"
	"errors"

	appusers "github.com/4nd3r5on/oidc-serv/internal/app/users"
	genapi "github.com/4nd3r5on/oidc-serv/pkg/api"
	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

func (h *Handlers) GetMe(ctx context.Context) (genapi.GetMeRes, error) {
	user, err := h.Me.Get(ctx)
	if err != nil {
		if errors.Is(err, errs.ErrUnauthorized) {
			return &genapi.ErrorResponse{Error: err.Error()}, nil
		}
		return nil, err
	}
	return toUser(user), nil
}

func (h *Handlers) UpdateMe(ctx context.Context, req *genapi.UpdateUserRequest) (genapi.UpdateMeRes, error) {
	err := h.Me.Update(ctx, updateOptsFromReq(req))
	if err != nil {
		if errors.Is(err, errs.ErrInvalidArgument) {
			return &genapi.UpdateMeBadRequest{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrUnauthorized) {
			return &genapi.UpdateMeUnauthorized{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrExists) {
			return conflictResponse(err), nil
		}
		return nil, err
	}
	return &genapi.UpdateMeNoContent{}, nil
}

func (h *Handlers) DeleteMe(ctx context.Context) (genapi.DeleteMeRes, error) {
	if err := h.Me.Delete(ctx); err != nil {
		if errors.Is(err, errs.ErrUnauthorized) {
			return &genapi.ErrorResponse{Error: err.Error()}, nil
		}
		return nil, err
	}
	return &genapi.DeleteMeNoContent{}, nil
}

func (h *Handlers) UpdateMyPassword(ctx context.Context, req *genapi.UpdatePasswordRequest) (genapi.UpdateMyPasswordRes, error) {
	err := h.Me.UpdatePassword(ctx, appusers.UpdatePasswordOpts{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidArgument) {
			return &genapi.UpdateMyPasswordBadRequest{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrUnauthorized) {
			return &genapi.UpdateMyPasswordUnauthorized{Error: err.Error()}, nil
		}
		return nil, err
	}
	return &genapi.UpdateMyPasswordNoContent{}, nil
}
