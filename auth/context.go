package auth

import (
	"golang.org/x/net/context"
)

type authKey struct{}

func FromContext(ctx context.Context) (Auth, bool) {
	c, ok := ctx.Value(authKey{}).(Auth)
	return c, ok
}

func NewContext(ctx context.Context, c Auth) context.Context {
	return context.WithValue(ctx, authKey{}, c)
}
