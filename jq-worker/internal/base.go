package internal

import (
	"context"
	"time"
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
	Register(ctx context.Context, info *WorkerInfo) error
	Enqueue(ctx context.Context, job *JobInfo) error
	Dequeue(ctx context.Context) (*JobInfo, error)

	Close() error
	FlushAll() error
}

type Message interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}

// WorkerInfo represents information about a worker.
type WorkerInfo struct {
	Id        string `msgpack:"id"`          // unique identifier for the worker (e.g., UUID v7)
	Name      string `msgpack:"worker_name"` // name of the worker (e.g., "worker-1")
	Processes int    `msgpack:"processes"`   // number of processes the worker can handle
}

// Encode encodes the WorkerInfo struct into a byte slice.
func (w *WorkerInfo) Encode() ([]byte, error) {
	return encodeMsg(w)
}

// Decode decodes the byte slice into a WorkerInfo struct.
func (w *WorkerInfo) Decode(data []byte) error {
	return decodeMsg(data, w)
}

type JobInfo struct {
	Id           string                 // unique identifier for the job (e.g., UUID v7)
	Name         string                 // name of the job (e.g., "job-1")
	Argument     map[string]interface{} // arguments for the job
	Priority     int                    // priority of the job (0 is the highest priority)
	MaxRetry     int                    // maximum number of retries for the job
	CurrentRetry int                    // current number of retries for the job
	KeepResult   bool                   // whether to keep the result of the job
	Timeout      time.Duration          // timeout for the job
	RegisteredAt time.Time              // time when the job was registered
}

// Encode encodes the JobInfo struct into a byte slice.
func (j *JobInfo) Encode() ([]byte, error) {
	return encodeMsg(j)
}

// Decode decodes the byte slice into a JobInfo struct.
func (j *JobInfo) Decode(data []byte) error {
	return decodeMsg(data, j)
}
