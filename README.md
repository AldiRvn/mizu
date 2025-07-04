# Mizu

> 💧 A flexible and SOLID-oriented utility toolkit for Go projects.  
> Inspired by the calm, adaptive nature of water — simple, smooth, and powerful.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/AldiRvn/mizu/actions/workflows/go.yml/badge.svg)](https://github.com/AldiRvn/mizu/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/AldiRvn/mizu)](https://goreportcard.com/report/github.com/AldiRvn/mizu)
[![Coverage Status](https://coveralls.io/repos/github/AldiRvn/mizu/badge.svg?branch=master)](https://coveralls.io/github/AldiRvn/mizu?branch=master)

## Setup

```bash
go get github.com/AldiRvn/mizu
```

```golang
import "github.com/AldiRvn/mizu"
```

## Example

```golang
redisPubSub := NewRedisLpushBrpop("localhost:6379", "", "0")
pubSub := NewPubSub(redisPubSub, redisPubSub) 

//? Use NewPubSub(redisPubSub, nil) if you only need to publish
//? Or NewPubSub(nil, redisPubSub) if you only need to subscribe

pubSub.Publish(ctx, redisKey, []byte(`{"a":"b"}`))

pubSub.Subcribe(ctx, redisKey, func(value []byte, err error) {
    if err != nil {
        slog.Error(err.Error())
        return
    }
    slog.Debug(string(value))
})

kafkaAddr := "localhost:29092"
kafkaTopic := "kafkaTopicTest"
kafkaUtil := NewKafkaPublish(kafkaAddr)
kafkaSub := NewPubSub(nil, NewKafkaConsumer(kafkaAddr, "kafkaTopicConsumer:0.0.1", kafka.FirstOffset))
kafkaPub := NewPubSub(kafkaUtil, nil)

kafkaPub.Publish(ctx, kafkaTopic, []byte("TWS"))

kafkaSub.Subcribe(ctx, kafkaTopic, func(value []byte, err error) { slog.Info("test subcribe 2", "value", string(value), "err", err) })
```
