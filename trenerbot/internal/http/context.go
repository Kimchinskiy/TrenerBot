package http

import (
	"context"

	"trenerbot/internal/domain"
)

type ctxKey int

const userCtxKey ctxKey = iota

func withUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

func UserFrom(ctx context.Context) *domain.User {
	u, _ := ctx.Value(userCtxKey).(*domain.User)
	return u
}
