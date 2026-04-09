package users

import (
	"context"

	"github.com/google/uuid"
)

// AuthFunc resolves the caller's identity from context.
// Matches the signature of [Authenticator.Auth].
type AuthFunc func(ctx context.Context, required bool) (uuid.UUID, error)

type (
	GetByIDFunc            func(ctx context.Context, id uuid.UUID) (*GetRes, error)
	UpdateFunc             func(ctx context.Context, id uuid.UUID, opts UpdateOpts) error
	DeleteFunc             func(ctx context.Context, id uuid.UUID) error
	UpdatePasswordByIDFunc func(ctx context.Context, id uuid.UUID, opts UpdatePasswordOpts) error
)

// Me composes self-service operations that always act on the authenticated caller.
// Auth is called on every method to resolve the user ID from context;
// the resolved ID is then forwarded to the relevant core operation.
type Me struct {
	Auth               AuthFunc
	GetCore            GetByIDFunc
	UpdateCore         UpdateFunc
	DeleteCore         DeleteFunc
	UpdatePasswordCore UpdatePasswordByIDFunc
}

func (me *Me) Get(ctx context.Context) (*GetRes, error) {
	userID, err := me.Auth(ctx, true)
	if err != nil {
		return nil, err
	}
	return me.GetCore(ctx, userID)
}

func (me *Me) UpdatePassword(ctx context.Context, opts UpdatePasswordOpts) error {
	userID, err := me.Auth(ctx, true)
	if err != nil {
		return err
	}
	return me.UpdatePasswordCore(ctx, userID, opts)
}

func (me *Me) Update(ctx context.Context, opts UpdateOpts) error {
	userID, err := me.Auth(ctx, true)
	if err != nil {
		return err
	}
	return me.UpdateCore(ctx, userID, opts)
}

func (me *Me) Delete(ctx context.Context) error {
	userID, err := me.Auth(ctx, true)
	if err != nil {
		return err
	}
	return me.DeleteCore(ctx, userID)
}
