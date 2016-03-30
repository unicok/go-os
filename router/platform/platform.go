package platform

import (
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-platform/router"
)

func NewRouter(opts ...selector.Option) router.Router {
	return router.NewRouter(opts...)
}
