package internal

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisOpt struct {
	Addr string
	User string
	Pass string
}

type RedisConn struct {
	client *redis.Client
}

func (r RedisConn) Self() *redis.Client {
	return r.client
}

func (r RedisConn) Register(ctx context.Context, msg *WorkerInfo) error {
	encMsg, err := msg.Encode()
	if err != nil {
		return err
	}
	return r.client.RPush(ctx, WorkerRegisterQueue, encMsg).Err()
}

func (r RedisConn) Enqueue() error {
	//TODO implement me
	panic("implement me")
}

func (r RedisConn) FlushAll() error {
	return r.client.FlushAll(context.Background()).Err()
}

func NewClient(redisOpt RedisOpt) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisOpt.Addr,
		Username: redisOpt.User,
		Password: redisOpt.Pass,
	})

	return &Client{Worker: RedisConn{client: client}}
}
