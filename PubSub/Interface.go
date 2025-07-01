package pubsub

import "context"

type Publisher interface {
	Publish(ctx context.Context, key string, value []byte) (err error)
}

type Subcriber interface {
	Subcribe(ctx context.Context, key string, callBack func(value []byte, err error))
}

type PubSub struct {
	pub Publisher
	sub Subcriber
}

func NewPubSub(pub Publisher, sub Subcriber) *PubSub {
	return &PubSub{
		pub: pub,
		sub: sub,
	}
}

func (ps *PubSub) Publish(ctx context.Context, key string, value []byte) (err error) {
	return ps.pub.Publish(ctx, key, value)
}

func (ps *PubSub) Subcribe(ctx context.Context, key string, callBack func(value []byte, err error)) {
	ps.sub.Subcribe(ctx, key, callBack)
}
