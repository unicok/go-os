package main

import (
	log "github.com/golang/glog"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/examples/server/handler"
	proto "github.com/micro/go-micro/examples/server/proto/example"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/server"

	"github.com/micro/go-platform/trace"
	"golang.org/x/net/context"
)

type Ping struct{}

func (p *Ping) Pong(ctx context.Context, req *proto.Ping, rsp *proto.Pong) error {
	rq := client.NewRequest("go.micro.srv.example", "Example.Call", &proto.Request{
		Name: "John",
	})

	rp := &proto.Response{}

	// Call service
	if err := client.Call(ctx, rq, rp); err != nil {
		return err
	}

	return nil
}

func main() {
	// optionally setup command line usage
	cmd.Init()

	t := trace.NewTrace()

	if err := t.Start(); err != nil {
		log.Fatal(err)
	}

	srv := &registry.Service{
		Name: "go.micro.srv.example",
	}

	client.DefaultClient = client.NewClient(
		client.Wrap(
			trace.ClientWrapper(t, srv),
		),
	)

	server.DefaultServer = server.NewServer(
		server.WrapHandler(trace.HandlerWrapper(t, srv)),
	)

	// Initialise Server
	server.Init(
		server.Name("go.micro.srv.example"),
	)

	// Register Handlers
	server.Handle(
		server.NewHandler(
			new(handler.Example),
		),
	)

	server.Handle(
		server.NewHandler(
			new(Ping),
		),
	)

	// Run server
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}

	if err := t.Stop(); err != nil {
		log.Fatal(err)
	}
}
