package internal

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	DefaultNumProcesses = 1
	DefaultPingInterval = 5 * time.Second
)

func genUUIDv7() string {
	v7, err := uuid.NewV7()
	if err != nil {
		return ""
	}
	return v7.String()
}

func getMachineHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

// ClientOption is an interface that defines the apply method
type ClientOption interface {
	apply(*WorkerInfo)
}

type hostnameOption string

func (h hostnameOption) apply(info *WorkerInfo) {
	info.Name = string(h)
}

func WithHostname(hostname string) ClientOption {
	return hostnameOption(hostname)
}

type processesOption int

func (p processesOption) apply(info *WorkerInfo) {
	info.Processes = int(p)
}

func WithProcesses(processes int) ClientOption {
	return processesOption(processes)
}

func NewClient(redisOpt RedisOpt, opts ...ClientOption) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisOpt.Addr,
		Username: redisOpt.User,
		Password: redisOpt.Pass,
	})

	info := WorkerInfo{
		Id:        genUUIDv7(),
		Name:      getMachineHostname(),
		Processes: DefaultNumProcesses,
	}

	for _, opt := range opts {
		opt.apply(&info)
	}

	rdbClient := Client{Worker: RedisConn{client}, Info: info}

	// send a ping message to the master
	go func() {
		ticker := time.NewTicker(DefaultPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := rdbClient.Ping(context.Background())
				if err != nil {
					continue
				}
			}
		}
	}()

	return &rdbClient
}

func (c *Client) FlushAll() error {
	return c.Worker.FlushAll()
}

func (c *Client) Close() error {
	return c.Worker.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Worker.Ping(ctx, c.Info.Id)
}

func (c *Client) Register(ctx context.Context) error {
	return c.Worker.Register(ctx, &c.Info)
}

// Enqueue **THIS IS A DEBUGGING FUNCTION**
func (c *Client) Enqueue(ctx context.Context, job *JobInfo) error {
	job.RegisteredAt = time.Now()
	return c.Worker.Enqueue(ctx, job)
}

func (c *Client) Dequeue(ctx context.Context) (*JobInfo, error) {
	return c.Worker.Dequeue(ctx)
}

func (c *Client) ReportResult(ctx context.Context, jobId string) error {
	return c.Worker.ReportResult(ctx, jobId)
}
