package transport

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type Conn struct {
	r *redis.Client
}

func NewConn(r *redis.Client) *Conn {
	return &Conn{r: r}
}

const (
	RGlobalQueue    = "jq:globalQueue"
	RWorkerRegister = "jq:worker-register"
	RResult         = "jq:result"
	RJobList        = "jq:jobList"
)

type WorkerRegisterRequest struct {
	ID         string `msgpack:"id"`
	WorkerName string `msgpack:"worker_name"`
	Processes  int    `msgpack:"processes"`
}

func (c *Conn) PollNewClient(ctx context.Context, workerChan chan<- WorkerRegisterRequest) {
	// Poll new client
	for {
		s, err := c.r.BLPop(ctx, 0, RWorkerRegister).Result()
		if err != nil {
			panic(err)
		}
		workerInfoPack := s[1]
		var workerInfo WorkerRegisterRequest
		if err = msgpack.Unmarshal([]byte(workerInfoPack), &workerInfo); err != nil {
			log.Printf("invalid worker registration request: %v", err)
			continue
		}

		log.Printf("worker registered: %v", workerInfo)
		workerChan <- workerInfo
	}
}

type JobRegisterRequest struct {
	ID         string                 `msgpack:"id"`
	Name       string                 `msgpack:"name"`
	Argument   map[string]interface{} `msgpack:"argument"`
	Priority   int                    `msgpack:"priority"`
	MaxRetry   int                    `msgpack:"max_retry"`
	KeepResult bool                   `msgpack:"keep_result"`
	Timeout    int                    `msgpack:"timeout"`
}

func (c *Conn) PollNewJob(ctx context.Context, jobChan chan<- JobRegisterRequest) {
	// poll new jobs
	for {
		s, err := c.r.BLPop(ctx, 0, RJobList).Result()
		if err != nil {
			panic(err)
		}
		jobPack := s[1]
		var job JobRegisterRequest
		if err = msgpack.Unmarshal([]byte(jobPack), &job); err != nil {
			log.Printf("invalid job registration request: %v", err)
			continue
		}

		log.Printf("job added: %v", job)
		jobChan <- job
	}
}
