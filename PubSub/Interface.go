package pubsub

import (
	"context"
	"errors"
	"log/slog"
)

type Publisher interface {
	Publish(ctx context.Context, key string, value []byte) (err error)
}

type Subscriber interface {
	Subscribe(ctx context.Context, key string, callBack func(value []byte, err error))
}

type PubSub struct {
	pub Publisher
	sub Subscriber
}

func NewPubSub(pub Publisher, sub Subscriber) *PubSub {
	return &PubSub{
		pub: pub,
		sub: sub,
	}
}

func (ps *PubSub) Publish(ctx context.Context, key string, value []byte) (err error) {
	if ps.pub == nil {
		err = errors.New("pub is nil")
		slog.Error("pub", "err", err)
		return
	}
	return ps.pub.Publish(ctx, key, value)
}

func (ps *PubSub) Subscribe(ctx context.Context, key string, callBack func(value []byte, err error)) {
	if ps.sub == nil {
		err := errors.New("sub is nil")
		slog.Error("sub", "err", err)
		callBack(nil, err)
		return
	}
	slog.Info("starting subscription", "key", key)
	ps.sub.Subscribe(ctx, key, callBack)
}
