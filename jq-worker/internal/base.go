package internal

import (
	"context"
	"time"
)

const (
	WorkerRegisterQueue = "jq:workerRegister"
	GlobalQueue         = "jq:globalQueue"
	ResultQueue         = "jq:resultQueue"
	JobRegisterQueue    = "jq:jobList"
)

type Client struct {
	Worker Worker
	Info   WorkerInfo
}

type Worker interface {
	Register(ctx context.Context, info *WorkerInfo) error
	Enqueue(ctx context.Context, job *JobInfo) error
	Dequeue(ctx context.Context) (*JobInfo, error)
	ReportResult(ctx context.Context, jobId string) error

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
	Id           string                 `msgpack:"id"`        // unique identifier for the job (e.g., UUID v7)
	Name         string                 `msgpack:"name"`      // name of the job (e.g., "job-1")
	Argument     map[string]interface{} `msgpack:"argument"`  // arguments for the job
	Priority     int                    `msgpack:"priority"`  // priority of the job (0 is the highest priority)
	MaxRetry     int                    `msgpack:"max_retry"` // maximum number of retries for the job
	CurrentRetry int                    // current number of retries for the job
	KeepResult   bool                   `msgpack:"keep_result"` // whether to keep the result of the job
	Timeout      time.Duration          `msgpack:"timeout"`     // timeout for the job
	RegisteredAt time.Time              // time when the job was registered
}

type JobResult struct {
	JobID      string `msgpack:"id"`
	Type       string `msgpack:"type"`
	FinishedAt string `msgpack:"finished_at"`

	// When type == JobResultSuccess
	Result map[string]interface{} `msgpack:"result"`

	// When type == JobResultFailure
	Reason      string      `msgpack:"reason"`
	ShouldRetry bool        `msgpack:"should_retry"`
	Error       interface{} `msgpack:"error,omitempty"`
	Message     string      `msgpack:"message"`
}

// Encode encodes the JobInfo struct into a byte slice.
func (j *JobInfo) Encode() ([]byte, error) {
	return encodeMsg(j)
}

// Decode decodes the byte slice into a JobInfo struct.
func (j *JobInfo) Decode(data []byte) error {
	return decodeMsg(data, j)
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
	Id         string          // This should be the same as the JobInfo ID
	Type       JobResultStatus // success or failure
	FinishedAt time.Time       // time when the job finished ISO8601

	Result map[string]interface{} // result of the job. Success -> result, Failure -> error

	FailureReason JobFailureReason // reason for failure
	ShouldRetry   bool             // whether the job should be retried
	ErrorMessage  string           // error message if the job failed (for Web UI)
}
