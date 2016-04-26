# Sync interface

The sync interface provides a way to coordinate across a number of nodes. This can be used 
for leadership election, membership consensus, etc. It's a building block for application 
synchronization.

We want the ability to choose between CP and AP where CP is useful for transactional and leader 
behaviour and AP for eventually consistent semantics.

```go
type Sync interface {
        // distributed lock interface
        Lock(...LockOption) (Lock, error)
        // leader election interface
        Leader(...LeaderOption) (Leader, error)
}

type Lock interface {
        Id() string
        Acquire() error
        Release() error
}

type Leader interface {
        // Returns the current leader
        Leader() (*registry.Node, error)
        // Elect self to become leader
        Elect() (Elected, error)
        // Returns the status of this node
        Status() (LeaderStatus, error)
}

type Elected interface {
        // Returns a channel which indicates
        // when the leadership is revoked
        Revoked() (chan bool, error)
        // Resign the leadership
        Resign() error
}
```

## Supported Backends

- Consul
