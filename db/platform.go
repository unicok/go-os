package db

import (
	"fmt"

	db "github.com/micro/db-srv/proto/db"
	"github.com/micro/go-micro/client"

	"golang.org/x/net/context"
)

type platform struct {
	opts Options
	c    db.DBClient
}

func newPlatform(opts ...Option) DB {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	if len(options.Database) == 0 {
		options.Database = DefaultDatabase
	}

	if len(options.Table) == 0 {
		options.Table = DefaultTable
	}

	return &platform{
		opts: options,
		c:    db.NewDBClient("go.micro.srv.db", options.Client),
	}
}

func protoToRecord(r *db.Record) Record {
	if r == nil {
		return nil
	}

	metadata := map[string]interface{}{}

	for k, v := range r.Metadata {
		metadata[k] = v
	}

	return &record{
		id:       r.Id,
		created:  r.Created,
		updated:  r.Updated,
		metadata: metadata,
		bytes:    []byte(r.Bytes),
	}
}

func recordToProto(r Record) *db.Record {
	if r == nil {
		return nil
	}

	md := map[string]string{}

	for k, v := range r.Metadata() {
		md[k] = fmt.Sprintf("%v", v)
	}

	return &db.Record{
		Id:       r.Id(),
		Created:  r.Created(),
		Updated:  r.Updated(),
		Metadata: md,
		Bytes:    string(r.Bytes()),
	}
}

func (p *platform) Close() error {
	return nil
}

func (p *platform) Init(opts ...Option) error {
	// No reinits
	return nil
}

func (p *platform) Options() Options {
	return p.opts
}

func (p *platform) Read(id string) (Record, error) {
	rsp, err := p.c.Read(context.TODO(), &db.ReadRequest{
		Database: &db.Database{
			Name:  p.opts.Database,
			Table: p.opts.Table,
		},
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return protoToRecord(rsp.Record), nil
}

func (p *platform) Create(r Record) error {
	_, err := p.c.Create(context.TODO(), &db.CreateRequest{
		Database: &db.Database{
			Name:  p.opts.Database,
			Table: p.opts.Table,
		},
		Record: recordToProto(r),
	})
	return err
}

func (p *platform) Update(r Record) error {
	_, err := p.c.Update(context.TODO(), &db.UpdateRequest{
		Database: &db.Database{
			Name:  p.opts.Database,
			Table: p.opts.Table,
		},
		Record: recordToProto(r),
	})
	return err
}

func (p *platform) Delete(id string) error {
	_, err := p.c.Delete(context.TODO(), &db.DeleteRequest{
		Database: &db.Database{
			Name:  p.opts.Database,
			Table: p.opts.Table,
		},
		Id: id,
	})
	return err
}

func (p *platform) Search(md Metadata, limit, offset int64) ([]Record, error) {
	metadata := map[string]string{}
	for k, v := range md {
		metadata[k] = fmt.Sprintf("%v", v)
	}

	rsp, err := p.c.Search(context.TODO(), &db.SearchRequest{
		Database: &db.Database{
			Name:  p.opts.Database,
			Table: p.opts.Table,
		},
		Metadata: metadata,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	var records []Record

	for _, r := range rsp.Records {
		records = append(records, protoToRecord(r))
	}

	return records, nil
}

func (p *platform) String() string {
	return "platform"
}
