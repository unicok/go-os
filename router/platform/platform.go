package platform

import (
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-os/router"
)

func NewRouter(opts ...selector.Option) router.Router {
	return router.NewRouter(opts...)
}
