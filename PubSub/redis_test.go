package pubsub

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

var (
	redisKey        = "tws.data"
	redisPubSub     = NewRedisLpushBrpop("localhost:6379", "", "0")
	redisPubInvalid = NewRedisLpushBrpop("localhost:6379", "", "No") //? Test not numeric
	redisSubInvalid = NewRedisLpushBrpop("localhost:6379", "", "69") //? Test out of range
	pubSub          = NewPubSub(redisPubSub, redisPubSub)
)

func init() {
	w := os.Stdout
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:   slog.LevelDebug,
			NoColor: !isatty.IsTerminal(w.Fd()),
		}),
	))
}

func Test_Publish_Redis(t *testing.T) {
	fmt.Println("Test_Publish_Redis()")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubSub.Publish(ctx, redisKey, []byte(`{"a":"b"}`))
	pubSub.Publish(ctx, redisKey, []byte(`{"a":"b1"}`))
	pubSub.Publish(ctx, redisKey, []byte(`{"a":"b2"}`))
}

func Test_Subcribe_Redis(t *testing.T) {
	fmt.Println("Test_Subcribe_Redis()")
	Test_Publish_Redis(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pubSub.Subcribe(ctx, redisKey, func(value []byte, err error) {
		if err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Debug(string(value))
	})

	//? Test Fail
	pubSub = NewPubSub(redisPubInvalid, redisSubInvalid)
	for i := 0; i < 3; i++ {
		switch i {
		case 0:
			pubSub = NewPubSub(redisPubInvalid, redisSubInvalid)
		case 1:
			pubSub = NewPubSub(nil, redisSubInvalid)
		case 2:
			pubSub = NewPubSub(redisPubInvalid, nil)
		}
		pubSub.Publish(ctx, redisKey, []byte(`{"a":"b"}`))
		pubSub.Subcribe(ctx, redisKey, func(value []byte, err error) {
			if err != nil {
				slog.Error(err.Error())
				return
			}
			slog.Debug(string(value))
		})
	}

	<-ctx.Done()
	time.Sleep(time.Second)
	log.Println(ctx.Err())
}
