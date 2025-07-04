package pubsub

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaWriterConn map[string]*kafka.Writer
	kafkaMutex      *sync.Mutex
)

func init() {
	kafkaWriterConn = make(map[string]*kafka.Writer)
	kafkaMutex = &sync.Mutex{}

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		for _, w := range kafkaWriterConn {
			if err := w.Close(); err != nil {
				slog.Error("kafka", "err", "failed to close writer:"+err.Error())
			}
		}
	}()
}

type Kafka struct {
	Addr string
}

func NewKafka(addr string) *Kafka {
	return &Kafka{
		Addr: addr,
	}
}

func (k *Kafka) genId(key string) (res string) {
	return strings.Join([]string{k.Addr, key}, "@")
}

func (k *Kafka) setupConn(key string) *kafka.Writer {
	idConn := k.genId(key)

	kafkaMutex.Lock()
	defer kafkaMutex.Unlock()

	if w, found := kafkaWriterConn[idConn]; found {
		return w
	}

	w := &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(k.Addr, ",")...),
		Topic:    key,
		Balancer: &kafka.LeastBytes{},
	}
	kafkaWriterConn[idConn] = w
	return w
}

func (k *Kafka) publish(ctx context.Context, key string, msgs []kafka.Message) (err error) {
	w := k.setupConn(key)

	if err = w.WriteMessages(ctx, msgs...); err != nil {
		slog.Error("kafka publish", "err", "failed to write messages: "+err.Error())
		return
	}
	slog.Debug("kafka publish", "status", "done", "key", key, "msgLen", len(msgs))
	return
}

func (k *Kafka) Publish(ctx context.Context, key string, value []byte) (err error) {
	msgs := []kafka.Message{{Value: value}}
	return k.publish(ctx, key, msgs)
}
