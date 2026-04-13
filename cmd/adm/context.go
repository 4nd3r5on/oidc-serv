package main

import (
	"context"
	"strings"
)

type ctxKey int

const (
	ctxAdminKey ctxKey = iota
	ctxBaseURL
	ctxCmdStack
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

// ctxWithCmdName appends name to the subcommand call stack stored in ctx.
func ctxWithCmdName(ctx context.Context, name string) context.Context {
	stack := cmdStackFromCtx(ctx)
	next := make([]string, len(stack)+1)
	copy(next, stack)
	next[len(stack)] = name
	return context.WithValue(ctx, ctxCmdStack, next)
}

func cmdStackFromCtx(ctx context.Context) []string {
	v, _ := ctx.Value(ctxCmdStack).([]string)
	return v
}

// cmdLine returns the full command path accumulated in ctx (e.g. "oidc-adm clients create").
func cmdLine(ctx context.Context) string {
	return strings.Join(cmdStackFromCtx(ctx), " ")
}
