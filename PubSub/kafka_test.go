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
	kafkaAddr   = "localhost:29092"
	kafkaTopic  = "tws"
	kafkaPub    = NewKafka(kafkaAddr)
	kafkaSub    = NewKafka(kafkaAddr)
	kafkaPubSub = NewPubSub(kafkaPub, nil)
)

func Test_Kafka_Publish(t *testing.T) {
	fmt.Println("Kafka_Publish()")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	kafkaPubSub.Publish(ctx, kafkaTopic, []byte("TWS"))
	kafkaPubSub.Publish(ctx, kafkaTopic, []byte("TWS2"))

	<-ctx.Done()
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)
	log.Println("Done")
}
