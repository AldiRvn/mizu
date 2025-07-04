package pubsub

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"testing"
	"time"
)

var (
	kafkaAddr       = "localhost:29092"
	kafkaTopic      = "kafkaTopicTest"
	kafkaUtil       = NewKafka(kafkaAddr)
	invalidKafkaPub = NewKafka("no")
	// kafkaSub        = NewPubSub(nil, kafkaUtil)
	kafkaPub = NewPubSub(kafkaUtil, nil)
)

func Test_Kafka_Publish(t *testing.T) {
	fmt.Println("Kafka_Publish()")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

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
