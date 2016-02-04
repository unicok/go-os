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
	Read(id string) (Record, error)
	Create(id string, v Record) error
	Update(id string, v Record) error
	Delete(id string) error
	Search(md Metadata, limit, offset int64) ([]Record, error)
	String() string
}

type Metadata map[string]interface{}

type Record interface {
	Id() string
	Metadata() Metadata
	Bytes() []byte
	Scan(v interface{}) error
}
```

## Supported Databases

- Platform
- Cassandra
- MariaDB
- Elasticsearch
