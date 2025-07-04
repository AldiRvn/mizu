package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaWriterConn map[string]*kafka.Writer
	kafkaReaderConn map[string]*kafka.Reader
	kafkaMutex      *sync.Mutex
)

func init() {
	kafkaWriterConn = make(map[string]*kafka.Writer)
	kafkaReaderConn = make(map[string]*kafka.Reader)
	kafkaMutex = &sync.Mutex{}

	go kafkaCloseConn()
}

func kafkaCloseConn() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit
	for _, w := range kafkaWriterConn {
		if err := w.Close(); err != nil {
			slog.Error("kafka", "err", "failed to close writer:"+err.Error())
		}
	}
	for _, c := range kafkaReaderConn {
		if err := c.Close(); err != nil {
			slog.Error("kafka", "err", "failed to close reader:"+err.Error())
		}
	}
}

type Kafka struct {
	addr        string
	groupId     string
	startOffset int64
}

func NewKafkaPublish(addr string) *Kafka {
	return &Kafka{
		addr: addr,
	}
}

func NewKafkaConsumer(addr, groupId string, startOffset int64) *Kafka {
	return &Kafka{
		addr:        addr,
		groupId:     groupId,
		startOffset: startOffset,
	}
}

func (k *Kafka) genConnId(key string) (res string) {
	return strings.Join([]string{k.addr, key}, "@")
}

func (k *Kafka) setupWriterConn(key string) *kafka.Writer {
	idConn := k.genConnId(key)

	kafkaMutex.Lock()
	defer kafkaMutex.Unlock()

	if w, found := kafkaWriterConn[idConn]; found {
		return w
	}

	w := &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(k.addr, ",")...),
		Topic:    key,
		Balancer: &kafka.LeastBytes{},
	}
	kafkaWriterConn[idConn] = w
	return w
}

func (k *Kafka) setupConsumerConn(key string) *kafka.Reader {
	idConn := k.genConnId(key)

	kafkaMutex.Lock()
	defer kafkaMutex.Unlock()

	if r, found := kafkaReaderConn[idConn]; found {
		return r
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(k.addr, ","),
		GroupID:     k.groupId,
		Topic:       key,
		StartOffset: k.startOffset,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
	})
	kafkaReaderConn[idConn] = r
	return r
}

func (k *Kafka) publish(ctx context.Context, key string, msgs []kafka.Message) (err error) {
	w := k.setupWriterConn(key)

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

func (k *Kafka) Subcribe(ctx context.Context, key string, callBack func(value []byte, err error)) {
	go func() {
		conn := k.setupConsumerConn(key)

		for {
			m, err := conn.ReadMessage(ctx)
			if err != nil {
				slog.Error("kafka subcribe", "err", err)
				callBack(nil, err)
				time.Sleep(time.Second)
				continue
			}
			slog.Debug("kafka subcribe", "detail", fmt.Sprintf(
				"message at topic/partition/offset %v/%v/%v: %s = %s", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value)),
			)
			callBack(m.Value, err)
		}
	}()
}
