package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/errors"
	"github.com/micro/go-platform/router"

	proto "github.com/micro/router-srv/proto/router"
)

var (
	service = "go.micro.srv.router"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func selector(id int, r router.Router) {
	// select the service
	next, err := r.Select(service)
	if err != nil {
		fmt.Println(id, "error selecting", err)
		return
	}

	for i := 1; i < 1000; i++ {
		// get a node
		node, err := next()
		if err != nil {
			fmt.Println(id, "error getting next", err)
			return
		}

		// make some request
		// client.Call(foo, request)
		req := client.NewRequest(service, "Router.Stats", &proto.StatsRequest{})

		var dur time.Duration

		// lets set an error
		if d := (rand.Int() % i); d == 0 {
			dur = time.Millisecond * time.Duration(rand.Int()%20)
			err = errors.InternalServerError(service, "err")
		} else if d == 1 {
			dur = time.Second * 5
			err = errors.New(service, "timed out", 408)
		} else {
			dur = time.Millisecond * time.Duration(rand.Int()%10)
			err = nil
		}

		// mark the result
		r.Mark(service, node, err)

		// record timing
		r.Record(req, node, dur, err)

		fmt.Println(id, "selected", node.Id)
		time.Sleep(time.Millisecond*10 + time.Duration(rand.Int()%10))
	}
}

func main() {
	cmd.Init()

	r := router.NewRouter()

	if err := r.Start(); err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 3; i++ {
		go selector(i, r)
	}

	selector(3, r)

	if err := r.Stop(); err != nil {
		fmt.Println(err)
		return
	}
}
