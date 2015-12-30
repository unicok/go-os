package main

import (
	log "github.com/golang/glog"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/examples/server/handler"
	"github.com/micro/go-micro/server"

	"github.com/micro/go-platform/trace"
)

func main() {
	// optionally setup command line usage
	cmd.Init()

	t := trace.NewTrace()

	if err := t.Start(); err != nil {
		log.Fatal(err)
	}

	server.DefaultServer = server.NewServer(
		server.WrapHandler(trace.HandlerWrapper(t, nil)),
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

	// Run server
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}

	if err := t.Stop(); err != nil {
		log.Fatal(err)
	}
}
