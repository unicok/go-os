package platform

import (
	"github.com/micro/go-os/db"
)

func NewDB(opts ...db.Option) db.DB {
	return db.NewDB(opts...)
}
