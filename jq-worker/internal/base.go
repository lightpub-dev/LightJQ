package internal

import (
	"context"
)

const (
	WorkerRegisterQueue = "jq:workerRegister"
	GlobalQueue         = "jq:globalQueue"
	ResultQueue         = "jq:resultQueue"
)

type Client struct {
	Worker Worker
	Info   WorkerInfo
}

type Worker interface {
	Close() error
	Register(ctx context.Context, info *WorkerInfo) error
	Enqueue() error

	FlushAll() error
	//Self() *redis.Client
}

type Message interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}

// WorkerInfo represents information about a worker.
type WorkerInfo struct {
	Message
	Id        string // unique identifier for the worker (e.g., UUID v7)
	Name      string // name of the worker (e.g., "worker-1")
	Processes int    // number of processes the worker can handle
}

// Encode encodes the WorkerInfo struct into a byte slice.
func (w *WorkerInfo) Encode() ([]byte, error) {
	return encodeMsg(w)
}

// Decode decodes the byte slice into a WorkerInfo struct.
func (w *WorkerInfo) Decode(data []byte) error {
	return decodeMsg(data, w)
}
