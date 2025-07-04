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
	kafkaAddr       = "localhost:29092"
	kafkaTopic      = "kafkaTopicTest"
	kafkaUtil       = NewKafkaPublish(kafkaAddr)
	invalidKafkaPub = NewKafkaPublish("no")
	kafkaSub        = NewPubSub(nil, NewKafkaConsumer(kafkaAddr, "kafkaTopicConsumer:0.0.1", kafka.FirstOffset))
	kafkaPub        = NewPubSub(kafkaUtil, nil)
)

func Test_Kafka(t *testing.T) {
	fmt.Println("Kafka_Publish()")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kafkaSub.Subcribe(ctx, kafkaTopic, func(value []byte, err error) { slog.Info("test subcribe 1", "value", string(value), "err", err) })
	kafkaSub.Subcribe(ctx, kafkaTopic, func(value []byte, err error) { slog.Info("test subcribe 2", "value", string(value), "err", err) })
	kafkaPub.Publish(ctx, kafkaTopic, []byte("TWS"))
	kafkaPub.Publish(ctx, kafkaTopic, []byte("TWS2"))

	//? Fail test
	kafkaPub.Publish(ctx, kafkaTopic, nil)
	kafkaPub.Subcribe(ctx, "", func(value []byte, err error) {})

	<-ctx.Done()
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)
	log.Println("Done")
}
