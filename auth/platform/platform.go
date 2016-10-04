package platform

import (
	"github.com/micro/go-os/auth"
)

func NewAuth(opts ...auth.Option) auth.Auth {
	return auth.NewAuth(opts...)
}
