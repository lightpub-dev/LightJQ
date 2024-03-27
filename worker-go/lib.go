package worker_go

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Worker struct {
	r          *redis.Client
	workerName string
	ctx        context.Context

	callbacks map[string]func(interface{}) error
}

type WorkerConfig struct{}

func NewWorker(config WorkerConfig) (*Worker, error) {
	// Create a new Worker instance
	return nil, nil
}
