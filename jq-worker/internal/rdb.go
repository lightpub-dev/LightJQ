package internal

import "github.com/redis/go-redis/v9"

type RedisOpt struct {
	Addr string
	User string
	Pass string
}

type RedisConn struct {
	client *redis.Client
}

func (r RedisConn) Enqueue() error {
	//TODO implement me
	panic("implement me")
}

func NewClient(redisOpt RedisOpt) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisOpt.Addr,
		Username: redisOpt.User,
		Password: redisOpt.Pass,
	})

	return &Client{worker: RedisConn{client: client}}
}
