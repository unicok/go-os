# Go Platform [![GoDoc](https://godoc.org/github.com/micro/go-platform?status.svg)](https://godoc.org/github.com/micro/go-platform) [![Travis CI](https://travis-ci.org/micro/go-platform.svg?branch=master)](https://travis-ci.org/micro/go-platform)

This is a microservice platform library built to be used with micro/go-micro. 
It provides all of the fundamental tooling required to run and manage 
a microservice environment. It's pluggable just like go-micro. It will be 
especially vital for anything above 20+ services and preferred by 
organisations. Developers looking to write standalone services should 
continue to use go-micro. 

Go platform depends on middleware/before-after funcs being added to 
go-micro. While each package can be used independently, it will be 
much more powerful as a whole.

Further community wide features should be contributed to [go-plugins](https://github.com/micro/go-plugins).

## Features
Each package provides a feature interface that will be pluggable and backed by a 
number of services.

Package     |   Built-in Plugin	|	Description
-------     |   --------	|	---------
[auth](https://godoc.org/github.com/micro/go-platform/auth)	|	auth-srv	|   authentication and authorisation for users and services	
[config](https://godoc.org/github.com/micro/go-platform/config)	|	config-srv	|   dynamic configuration which is namespaced and versioned
[db](https://godoc.org/github.com/micro/go-platform/db)		|	db-srv		| distributed database abstraction
[discovery](https://godoc.org/github.com/micro/go-platform/discovery)	|	discovery-srv	|   extends the go-micro registry to add heartbeating, etc
[event](https://godoc.org/github.com/micro/go-platform/event)	|	event-srv	|	platform event publication, subscription and aggregation 
[kv](https://godoc.org/github.com/micro/go-platform/kv)		|	distributed in-memory	|   simply key value layered on memcached, etcd, consul 
[log](https://godoc.org/github.com/micro/go-platform/log)	|	file	|	structured logging to stdout, logstash, fluentd, pubsub
[monitor](https://godoc.org/github.com/micro/go-platform/monitor)	|	monitor-srv	|   add custom healthchecks measured with distributed systems in mind
[metrics](https://godoc.org/github.com/micro/go-platform/metrics)	|	telegraf	|   instrumentation and collation of counters
[router](https://godoc.org/github.com/micro/go-platform/router)	|	router-srv	|	global circuit breaking, load balancing, A/B testing
[sync](https://godoc.org/github.com/micro/go-platform/sync)	|	consul		|	distributed locking, leadership election, etc
[trace](https://godoc.org/github.com/micro/go-platform/trace)	|	trace-srv	|	distributed tracing of request/response

## What's it even good for?

The Micro platform is useful for where you want to build a reliable globally distributed systems platform at scale. 
You would be in good company by doing so, with the likes of Google, Facebook, Amazon, Twitter, Uber, Hailo, etc, etc.

![Micro On-Demand](https://github.com/micro/micro/blob/master/doc/ondemand.png)

## How does it work?

The go-platform is a client side interface for the fundamentals of a microservice platform. Each package connects to 
a service which handles that feature. Everything is an interface and pluggable which means you can choose how to 
architect your platform. Micro however provides a "platform" implementation backed by it's own services by default.

