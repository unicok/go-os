# Config - Dynamic config interface

Provides a high level pluggable abstraction for dynamic configuration.

## Interface

There's a need for dynamic configuration with namespacing, deltas for rollback, 
watches for changes and an audit log. At a low level we may care about server 
addresses changing, routing information, etc. At a high level there may be a 
need to control business level logic; External API Urls, Pricing information, etc.

```
config.New(namespace) // namespaced by service name
config.Load(namespace) // load existing namespace
config.Watch(path) // /[service]/api/urls
```

## Supported Backends

- Cassandra
- Zookeeper/Etcd
- Consul
