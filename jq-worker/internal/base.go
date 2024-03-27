package internal

const (
	WorkerRegisterQueue = "jq:workerRegister"
	GlobalQueue         = "jq:globalQueue"
	ResultQueue         = "jq:resultQueue"
)

type Client struct {
	worker Worker
}

type Worker interface {
	Enqueue() error
}
