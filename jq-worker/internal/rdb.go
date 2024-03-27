package internal

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// RedisOpt is a struct that contains the address, username, and password of a Redis instance.
type RedisOpt struct {
	Addr string
	User string
	Pass string
}

// RedisConn is a struct that holds a connection to a Redis instance
//
// implements the Worker interface
type RedisConn struct {
	Client *redis.Client
}

func (r RedisConn) Close() error {
	return r.Client.Close()
}

func (r RedisConn) Self() *redis.Client {
	return r.Client
}

func (r RedisConn) Register(ctx context.Context, info *WorkerInfo) error {
	encMsg, err := info.Encode()
	if err != nil {
		return err
	}
	return r.Client.RPush(ctx, WorkerRegisterQueue, encMsg).Err()
}

func (r RedisConn) Enqueue() error {
	//TODO implement me
	panic("implement me")
}

func (r RedisConn) FlushAll() error {
	return r.Client.FlushAll(context.Background()).Err()
}
