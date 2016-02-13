package main

import (
	"fmt"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	example "github.com/micro/go-micro/examples/server/proto/example"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-platform/trace"
	"golang.org/x/net/context"
	"time"
)

func call(i int) {
	// Create new request to service go.micro.srv.example, method Example.Call
	req := client.NewRequest("go.micro.srv.example", "Example.Call", &example.Request{
		Name: "John",
	})

	// create context with metadata
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"X-User-Id": "john",
		"X-From-Id": "script",
	})

	rsp := &example.Response{}

	// Call service
	if err := client.Call(ctx, req, rsp); err != nil {
		fmt.Println("call err: ", err, rsp)
		return
	}

	fmt.Println("Call:", i, "rsp:", rsp.Msg)
}

func main() {
	cmd.Init()

	t := trace.NewTrace()

	srv := &registry.Service{
		Name: "go.client",
	}

	client.DefaultClient = client.NewClient(
		client.Wrap(
			trace.ClientWrapper(t, srv),
		),
	)

	fmt.Println("Starting trace")
	if err := t.Start(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("\n--- Traced Call example ---\n")
	i := 0
	for {
		call(i)
		i++
		<-time.After(time.Second * 5)
	}

	if err := t.Stop(); err != nil {
		fmt.Println(err)
	}
}
