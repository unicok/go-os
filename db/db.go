package db

type DB interface {
	Init(...Option) error
	Read(id string) (Record, error)
	Create(id string, v Record) error
	Update(id string, v Record) error
	Delete(id string) error
	Search(md Metadata) ([]Record, error)
	String() string
}

type Option func(*Options)

type Metadata map[string]interface{}

type Record interface {
	Id() string
	Metadata() Metadata
	Bytes() []byte
	Scan(v interface{}) error
}
