package auth

import (
	"context"

	"github.com/google/uuid"
)

type (
	ctxClientDataKey    struct{}
	ctxClientDataSetKey struct{}
	ctxUserIDKey        struct{}
)

func CtxGetClientData(ctx context.Context) (*ClientData, bool) {
	clientDataSet, _ := ctx.Value(ctxClientDataSetKey{}).(bool)
	if !clientDataSet {
		return nil, false
	}
	clientData, valid := ctx.Value(ctxClientDataKey{}).(*ClientData)
	return clientData, valid
}

func CtxPutClientData(ctx context.Context, clientData *ClientData) context.Context {
	ctx = context.WithValue(ctx, ctxClientDataSetKey{}, true)
	if clientData != nil {
		ctx = context.WithValue(ctx, ctxClientDataKey{}, clientData)
	}
	return ctx
}

func CtxGetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxUserIDKey{}).(uuid.UUID)
	return id, ok
}

func CtxPutUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxUserIDKey{}, userID)
}
