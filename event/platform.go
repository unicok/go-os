package event

import (
	"time"

	ev "github.com/micro/event-srv/proto/event"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-platform/event/proto"
	"github.com/pborman/uuid"

	"golang.org/x/net/context"
)

type platform struct {
	opts Options

	cl ev.EventClient
}

func newPlatform(opts ...Option) Event {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	return &platform{
		opts: options,
		cl:   ev.NewEventClient("go.micro.srv.event", options.Client),
	}
}

func toRecord(r *event.Record) *Record {
	return &Record{
		Id:        r.Id,
		Type:      r.Type,
		Origin:    r.Origin,
		Timestamp: r.Timestamp,
		RootId:    r.RootId,
		Metadata:  r.Metadata,
		Data:      r.Data,
	}
}

func toProto(r *Record) *event.Record {
	return &event.Record{
		Id:        r.Id,
		Type:      r.Type,
		Origin:    r.Origin,
		Timestamp: r.Timestamp,
		RootId:    r.RootId,
		Metadata:  r.Metadata,
		Data:      r.Data,
	}
}

func (p *platform) Publish(ctx context.Context, r *Record) error {
	if len(r.Type) == 0 {
		r.Type = DefaultEventType
	}

	if r.Timestamp == 0 {
		r.Timestamp = time.Now().Unix()
	}

	if len(r.Id) == 0 {
		r.Id = uuid.NewUUID().String()
	}

	pub := p.opts.Client.NewPublication(RecordTopic, toProto(r))
	return p.opts.Client.Publish(ctx, pub)
}

// currently blocking
func (p *platform) Subscribe(ctx context.Context, h Handler, types ...string) error {
	req := &ev.StreamRequest{}
	for _, typ := range types {
		req.Types = append(req.Types, typ)
	}

	stream, err := p.cl.Stream(ctx, req)
	if err != nil {
		return err
	}

	for {
		rsp, err := stream.Recv()
		if err != nil {
			return err
		}
		h(toRecord(rsp.Record))
	}

	return nil
}

func (p *platform) Start() error {
	return nil
}

func (p *platform) Stop() error {
	return nil
}

func (p *platform) String() string {
	return "platform"
}
