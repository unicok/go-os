# Go Platform [![GoDoc](https://godoc.org/github.com/micro/go-platform?status.svg)](https://godoc.org/github.com/micro/go-platform) [![Travis CI](https://travis-ci.org/micro/go-platform.svg?branch=master)](https://travis-ci.org/micro/go-platform) [![Go Report Card](https://goreportcard.com/badge/micro/go-platform)](https://goreportcard.com/report/github.com/micro/go-platform)

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

Examples can be found in [go-platform/examples](https://github.com/micro/go-platform/tree/master/examples)

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

Each package can be used independently or integrated using go-micro client and handler wrappers.

### Auth 

Auth addresses authentication and authorization of services and users. The default implementation is Oauth2 with an additional policy 
engine coming soon. This is the best way to authenticate users and service to service calls using a centralised 
authority. Security is a first class citizen in a microservice platform.

### Config 

Config implements an interface for dynamic configuration. The config can be hierarchically loaded and merged from 
multiple sources e.g file, url, config service. It can and should also be namespaced so that environment specific 
config is loaded when running in dev, staging or production. The config interface is useful for business level 
configuration required by your services. It can be reloaded without needing to restart a service.

### DB (experimental) 

The DB interface is an experiment CRUD interface to simplify database access and management. The amount of CRUD boilerplate 
written and rewritten in a microservice world is immense. By offloading this to a backend service and using RPC, we 
eliminate much of that and speed up development. The platform implementation includes pluggable backends such as mysql, 
cassandra, elasticsearch and utilises the registry to lookup which nodes databases are assigned to. 

This is purely experimental at this point based on some ideas from how Google, Facebook and Twitte do database management 
internally.
 
### Discovery 

Discovery provides a high level service discovery interface on top of the go-micro registry. It utilises the watcher to 
locally cache service records and also heartbeats to a discovery service. It's akin to the Netflix Eureka 2.0 
architecture whereby we split the read and write layers of discovery into separate services.

### Event

The event package provides a way to send platform events and essentially create an event stream and record of all that's 
happening in your microservice environment. On the backend an event service aggregates the records and allows you to 
subscribe to a specific set of events. An event driven architecture is a powerful concept in a microservice environment 
and must be addressed adequately. At scale it's essential for correlating events within a distributed system e.g 
provisioning of new services, change of dynamic config, logouts for customers, tracking notifications, alerts.
 
### KV 

KV represents a simple distributed key-value interface. It's useful for sharing small fast access bits of data amonst 
instances of a service. We provide three implementations currently. Memcached, redis and a consistently hashed in distributed 
in memory system.

### Log 

Log provides a structured logging interface which allows log messages to be tagged with key-value pairs. 
The default output plugin is file which allows many centralised logging systems to be used such as the ELK stack. 

### Monitor 

The monitor provides a way to publish Status, Stats and Healtchecks to a monitoring service. Healthchecks are user defined 
checks that may be critical to a service e.g can access database, can sync from s3, etc. Monitoring in a distributed 
system is fundamentally different from the classic LAMP stack. In the old ways pings and tcp checks were regarded as enough, 
in a distributed system we require much more fine grained metrics and a monitoring service which can make sense of what 
failure means in this world.

### Metrics 

Metrics is an interface for instrumentation. We regard metrics as a superior form of observability in a distributed system over 
logging. Instrumentation is a great way to graph historic and realtime data which can be correlated and immediately 
understood. The metrics interface provides a way to create counters, gauges and histograms. We currently implement the statsd 
interface and offload to telegraf which provides an augmented statsd interface with labels.

### Router

The router builds on the registry and selector to provide rate limiting, circuit breaking and global service load balancing. 
It implements the selector interface. Stats are recorded for every request and periodically published. A centralised routing 
service aggregates these metrics from all services in the environment and makes decisions about how to route requests. 
The routing service is not a proxy. Proxies are a weak form of load balancing, we prefer smart clients which retrieve 
a list of nodes from the router and make direct connections, this means if the routing service dies or misbehaves, clients 
can continue to make request independently.
 
### Sync 

Sync is an interface for distributed synchronisation. This provides an easy way to do leadership election and locking to 
serialise access to a resource. We expect there to be multiple copies of a service running to provide fault tolerance and 
scalability but it makes it much harder to deal with transactions or serialising access. The sync package provides a 
way to regain some of these semantics.
 
### Trace 

Trace is a client side interface for distributed tracing e.g dapper, zipkin, appdash. In a microservice world, a single 
request may fan out to 20-30 services. Failure may be non deterministic and difficult to track. Distributed tracing is a 
way of tracking the lifetime of a request. The interface utilises client and server wrappers to simplify using tracing.
