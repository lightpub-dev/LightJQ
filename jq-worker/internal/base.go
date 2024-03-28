package internal

import (
	"context"
	"time"
)

const (
	PingChannel         = "jq:ping"
	WorkerRegisterQueue = "jq:workerRegister"
	JobQueuePrefix      = "jq:job:"
	GlobalQueue         = "jq:globalQueue"
	ResultQueue         = "jq:resultQueue"
	JobRegisterQueue    = "jq:jobList"
	ProcessingSet       = "jq:processing"
)

type Client struct {
	Worker Worker
	Info   WorkerInfo
}

type Worker interface {
	Ping(ctx context.Context, workerID string) error
	Register(ctx context.Context, info *WorkerInfo) error
	Enqueue(ctx context.Context, job *JobInfo) error
	Dequeue(ctx context.Context) (*JobInfo, error)
	ReportResult(ctx context.Context, job *JobInfo) error

	Close() error
	FlushAll() error
}

type Message interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}

// PingMessage represents a ping message to notify master that the worker is alive.
type PingMessage struct {
	WorkerID string `msgpack:"worker_id"`
}

// Encode encodes the PingMessage struct into a byte slice.
func (p *PingMessage) Encode() ([]byte, error) {
	return encodeMsg(p)
}

// Decode decodes the byte slice into a PingMessage struct.
func (p *PingMessage) Decode(data []byte) error {
	return decodeMsg(data, p)
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
	Id           string                 `msgpack:"id"`        // unique identifier for the job (e.g., UUID v7)
	Name         string                 `msgpack:"name"`      // name of the job (e.g., "job-1")
	Argument     map[string]interface{} `msgpack:"argument"`  // arguments for the job
	Priority     int                    `msgpack:"priority"`  // priority of the job (0 is the highest priority)
	MaxRetry     int                    `msgpack:"max_retry"` // maximum number of retries for the job
	CurrentRetry int                    // current number of retries for the job
	KeepResult   bool                   `msgpack:"keep_result"` // whether to keep the result of the job
	Timeout      time.Duration          `msgpack:"timeout"`     // timeout for the job
	RegisteredAt string                 // time when the job was registered
	StartedAt    string                 // time when the job was started
}

// GenerateProcessingInfo generates a ProcessingInfo struct from the JobInfo struct.
func (j *JobInfo) GenerateProcessingInfo() ProcessingInfo {
	return ProcessingInfo{
		JobID:     j.Id,
		StartedAt: j.StartedAt,
		Timeout:   j.Timeout,
	}
}

// Encode encodes the JobInfo struct into a byte slice.
func (j *JobInfo) Encode() ([]byte, error) {
	return encodeMsg(j)
}

// Decode decodes the byte slice into a JobInfo struct.
func (j *JobInfo) Decode(data []byte) error {
	return decodeMsg(data, j)
}

type ProcessingInfo struct {
	JobID     string        `msgpack:"job_id"`
	StartedAt string        `msgpack:"started_at"`
	Timeout   time.Duration `msgpack:"timeout"`
}

// Encode encodes the ProcessingInfo struct into a byte slice.
func (p *ProcessingInfo) Encode() ([]byte, error) {
	return encodeMsg(p)
}

// Decode decodes the byte slice into a ProcessingInfo struct.
func (p *ProcessingInfo) Decode(data []byte) error {
	return decodeMsg(data, p)
}

type JobResultStatus string

const (
	JobResultStatusSuccess JobResultStatus = "success"
	JobResultStatusFailure JobResultStatus = "failure"
)

type JobFailureReason string

const (
	JobFailureReasonTimeout     JobFailureReason = "timeout"
	JobFailureReasonServerIssue JobFailureReason = "server_issue"
	JobFailureReasonUnknown     JobFailureReason = "unknown"
)

type JobResult struct {
	JobID      string          `msgpack:"id"`
	Type       JobResultStatus `msgpack:"type"`
	FinishedAt string          `msgpack:"finished_at"`

	// When type == JobResultSuccess
	Result map[string]interface{} `msgpack:"result"`

	// When type == JobResultFailure
	Reason      JobFailureReason `msgpack:"reason"`
	ShouldRetry bool             `msgpack:"should_retry"`
	Error       interface{}      `msgpack:"error,omitempty"`
	Message     string           `msgpack:"message"`
}
