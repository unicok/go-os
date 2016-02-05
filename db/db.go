package db

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

var (
	DefaultDatabase = "micro"
	DefaultTable    = "micro"
)

func NewDB(opts ...Option) DB {
	return newPlatform(opts...)
}

func NewRecord(id string, md Metadata, data interface{}) Record {
	return newRecord(id, md, data)
}
