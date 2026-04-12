package main

import "context"

type ctxKey int

const (
	ctxAdminKey ctxKey = iota
	ctxBaseURL
)

func ctxWithConfig(ctx context.Context, adminKey, baseURL string) context.Context {
	ctx = context.WithValue(ctx, ctxAdminKey, adminKey)
	return context.WithValue(ctx, ctxBaseURL, baseURL)
}

func adminKeyFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ctxAdminKey).(string)
	return v
}

func baseURLFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ctxBaseURL).(string)
	return v
}
