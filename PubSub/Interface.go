package pubsub

import (
	"context"
	"errors"
	"log/slog"
)

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
	if ps.pub == nil {
		err = errors.New("pub is nil")
		slog.Error("pub", "err", err)
		return
	}
	return ps.pub.Publish(ctx, key, value)
}

func (ps *PubSub) Subcribe(ctx context.Context, key string, callBack func(value []byte, err error)) {
	if ps.sub == nil {
		err := errors.New("sub is nil")
		slog.Error("sub", "err", err)
		callBack(nil, err)
		return
	}
	slog.Info("starting subscription", "key", key)
	ps.sub.Subcribe(ctx, key, callBack)
}
