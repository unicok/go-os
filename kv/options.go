package kv

import (
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"

	"golang.org/x/net/context"
)

type Options struct {
	// Servers that store values for access
	Servers []string
	// Limit the scope of queries to a namespace
	// In the case of platform it essentially
	// means the topic we pub to for "gossip"
	// For memcache and redis it could be key/prefix
	Namespace string
	// Number of replicas
	// Default is 1 ofcourse
	Replicas int

	// Replace with go-micro.Service?
	Client client.Client
	Server server.Server

	// Alternative options set using Context
	Context context.Context
}

func Servers(s []string) Option {
	return func(o *Options) {
		o.Servers = s
	}
}

func Namespace(ns string) Option {
	return func(o *Options) {
		o.Namespace = ns
	}
}

func Client(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

func Server(s server.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}
