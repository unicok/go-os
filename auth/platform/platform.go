package platform

import (
	"github.com/micro/go-platform/auth"
)

func NewAuth(opts ...auth.Option) auth.Auth {
	return auth.NewAuth(opts...)
}
