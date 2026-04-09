package api

import (
	"context"
	"errors"
	"net/http"

	appusers "github.com/4nd3r5on/oidc-serv/internal/app/users"
	genapi "github.com/4nd3r5on/oidc-serv/pkg/api"
	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

func (h *Handlers) CreateUser(ctx context.Context, req *genapi.CreateUserRequest) (genapi.CreateUserRes, error) {
	id, err := h.Create.Create(ctx, appusers.CreateOpts{
		Username: req.Username,
		Password: req.Password,
		Locale:   req.Locale.Or(""),
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidArgument) {
			return &genapi.CreateUserBadRequest{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrUnauthorized) {
			return &genapi.CreateUserUnauthorized{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrExists) {
			return conflictResponse(err), nil
		}
		return nil, err
	}
	return &genapi.CreateUserResponse{ID: id}, nil
}

func (h *Handlers) GetUserById(ctx context.Context, params genapi.GetUserByIdParams) (genapi.GetUserByIdRes, error) {
	user, err := h.GetByID.Get(ctx, params.UserId)
	if err != nil {
		if errors.Is(err, errs.ErrUnauthorized) {
			return &genapi.GetUserByIdUnauthorized{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrNotFound) {
			return &genapi.GetUserByIdNotFound{Error: err.Error()}, nil
		}
		return nil, err
	}
	return toUser(user), nil
}

func (h *Handlers) GetUserByUsername(ctx context.Context, params genapi.GetUserByUsernameParams) (genapi.GetUserByUsernameRes, error) {
	user, err := h.GetByUsername.Get(ctx, params.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUnauthorized) {
			return &genapi.GetUserByUsernameUnauthorized{Error: err.Error()}, nil
		}
		if errors.Is(err, errs.ErrNotFound) {
			return &genapi.GetUserByUsernameNotFound{Error: err.Error()}, nil
		}
		return nil, err
	}
	return toUser(user), nil
}

// NewError maps unhandled handler errors to the default internal error response.
func (h *Handlers) NewError(_ context.Context, err error) *genapi.InternalErrorStatusCode {
	return &genapi.InternalErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response:   genapi.ErrorResponse{Error: err.Error()},
	}
}
