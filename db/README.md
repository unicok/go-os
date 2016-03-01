# DB - database interface

Provides a high level pluggable abstraction for databases.

**This might be removed**

**Don't use this. Run away. RUN AWAY**

## Interface

Initial thoughts lie around a CRUD interface. The number of times 
one has to write CRUD on top of database libraries, having to think 
through schema and data modelling based on different databases is a 
pain. Going lower level than this doesn't pose any value.

Event sourcing can be tackled in a separate package.

```go
type DB interface {
        Init(...Option) error
        Options() Options
        Read(id string) (Record, error)
        Create(r Record) error
        Update(r Record) error
        Delete(id string) error
        Search(md Metadata, limit, offset int64) ([]Record, error)
        String() string
        // Potential expiremental in memory DB
        // needs to be started/stopped to be part of a ring
        // Also publishing to notify this is querying dbs
        Start() error
        Stop() error
}

type Option func(*Options)

type Metadata map[string]interface{}

type Record interface {
        Id() string
        Created() int64
        Updated() int64
        Metadata() Metadata
        Bytes() []byte
        Scan(v interface{}) error
}

func NewDB(opts ...Option) DB {
	return newPlatform(opts...)
}

func NewRecord(id string, md Metadata, data interface{}) Record {
	return newRecord(id, md, data)
}
```

## Supported Backends

- DB Service
