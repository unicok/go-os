# Router interface

The router is a client interface to a global or federated regional router. 
A go-micro client uses the registry and then on an individual basis calculates 
how to route. The router allows decisions to be made across the platform.

## Interface

```go
type Router interface {
	selector.Selector
	Start() error
	Stop() error
}
```
