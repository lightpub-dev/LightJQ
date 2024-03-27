package internal

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func NewClient(redisOpt RedisOpt) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisOpt.Addr,
		Username: redisOpt.User,
		Password: redisOpt.Pass,
	})

	info := WorkerInfo{
		Id:        "1",
		Name:      "worker-1",
		Processes: 1,
	}

	return &Client{Worker: RedisConn{client}, Info: info}
}

func (c *Client) FlushAll() error {
	return c.Worker.FlushAll()
}

func (c *Client) Close() error {
	return c.Worker.Close()
}

func (c *Client) Register(ctx context.Context) error {
	return c.Worker.Register(ctx, &c.Info)
}

func (c *Client) Enqueue() error {
	return c.Worker.Enqueue()
}
