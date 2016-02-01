package main

import (
	"fmt"
	"time"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-platform/router"
)

func selector(id int, r router.Router) {
	next, err := r.Select("go.micro.srv.router")
	if err != nil {
		fmt.Println(id, "error selecting", err)
		return
	}

	for i := 0; i < 10; i++ {
		node, err := next()
		if err != nil {
			fmt.Println(id, "error getting next", err)
			return
		}
		fmt.Println(id, "selected", node.Id)
		time.Sleep(time.Millisecond * 10)
	}
}

func main() {
	cmd.Init()

	r := router.NewRouter()

	for i := 0; i < 3; i++ {
		go selector(i, r)
	}

	selector(3, r)
}
