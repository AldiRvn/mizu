package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisConn  map[string]*redis.Client
	redisMutex *sync.Mutex
)

func init() {
	redisConn = make(map[string]*redis.Client)
	redisMutex = &sync.Mutex{}
}

type RedisLpushBrpop struct {
	Addr     string
	Password string
	Db       string
}

func NewRedisLpushBrpop(addr, password, db string) *RedisLpushBrpop {
	return &RedisLpushBrpop{
		Addr:     addr,
		Password: password,
		Db:       db,
	}
}

func (r *RedisLpushBrpop) getDb() (res int64, err error) {
	res, err = strconv.ParseInt(r.Db, 10, 64)
	if err != nil {
		slog.Error("redis get db", "err", err.Error())
	}
	return
}

func (r *RedisLpushBrpop) ping(ctx context.Context, rdb *redis.Client) error {
	keyConn := r.getKeyConn()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		slog.Error("redis ping failed", "addr/db", keyConn)
		return fmt.Errorf("failed to connect: %w", err)
	}
	slog.Debug("redis ping success", "addr/db", keyConn, "resp", pong)
	return nil
}

func (r *RedisLpushBrpop) getKeyConn() string {
	return fmt.Sprintf("%s/%s", r.Addr, r.Db)
}

func (r *RedisLpushBrpop) saveConn(ctx context.Context, rdb *redis.Client) error {
	redisMutex.Lock()
	defer redisMutex.Unlock()

	keyConn := r.getKeyConn()
	if _, found := redisConn[keyConn]; !found {
		if err := r.ping(ctx, rdb); err != nil {
			return err
		}

		redisConn[keyConn] = rdb
		slog.Debug("redis connected", "addr/db", keyConn)
	}
	return nil
}

func (r *RedisLpushBrpop) getConn() *redis.Client {
	if conn, found := redisConn[r.getKeyConn()]; found {
		return conn
	}
	return nil
}

func (r *RedisLpushBrpop) connect(ctx context.Context) (rdb *redis.Client, err error) {
	rdb = r.getConn()
	if rdb != nil {
		return
	}

	db, err := r.getDb()
	if err != nil {
		return nil, err
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       int(db),
	})

	if err := r.saveConn(ctx, rdb); err != nil {
		return nil, err
	}

	return rdb, nil
}

func (r *RedisLpushBrpop) Publish(ctx context.Context, key string, value []byte) (err error) {
	rdb, err := r.connect(ctx)
	if err != nil {
		return err
	}

	if err := rdb.LPush(ctx, key, value).Err(); err != nil {
		slog.Error("redis lpush", "err", err)
		return err
	}
	return nil
}

func (r *RedisLpushBrpop) Subscribe(ctx context.Context, key string, callBack func([]byte, error)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("redis stop subscribe to", "key", key)
				return
			default:
				rdb, err := r.connect(ctx)
				if err != nil {
					callBack(nil, err)
					time.Sleep(time.Second)
					continue
				}

				result, err := rdb.BRPop(ctx, 0, key).Result()
				if err != nil {
					callBack(nil, err)
					continue
				}

				if len(result) != 2 {
					callBack(nil, fmt.Errorf("redis unexpected BRPOP result"))
					continue
				}

				callBack([]byte(result[1]), nil)
			}
		}
	}()
}
