package pubsub

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaAddr  = "localhost:29092"
	kafkaTopic = "kafkaTopicTest"
	kafkaUtil  = NewKafkaPublish(kafkaAddr)
	kafkaSub   = NewPubSub(nil, NewKafkaConsumer(kafkaAddr, "kafkaTopicConsumer:0.0.1", kafka.FirstOffset))
	kafkaPub   = NewPubSub(kafkaUtil, nil)
)

func Test_Kafka(t *testing.T) {
	fmt.Println("Kafka_Publish()")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kafkaSub.Subscribe(ctx, kafkaTopic, func(value []byte, err error) { slog.Info("test subscribe 1", "value", string(value), "err", err) })
	kafkaSub.Subscribe(ctx, kafkaTopic, func(value []byte, err error) { slog.Info("test subscribe 2", "value", string(value), "err", err) })
	kafkaPub.Publish(ctx, kafkaTopic, []byte("TWS"))
	kafkaPub.Publish(ctx, kafkaTopic, []byte("TWS2"))

	//? Fail test
	kafkaPub.Publish(ctx, kafkaTopic, nil)
	kafkaPub.Subscribe(ctx, "", func(value []byte, err error) {})

	<-ctx.Done()
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)
	log.Println("Done")
}
