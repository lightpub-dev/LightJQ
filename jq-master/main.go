package main

import (
	"context"
	"log"
	"time"

	"github.com/lightpub-dev/lightjq/jq-master/scheduler"
	"github.com/lightpub-dev/lightjq/jq-master/transport"
	"github.com/redis/go-redis/v9"
)

type JQMaster struct {
	r *redis.Client

	conn  *transport.Conn
	sched *scheduler.Scheduler
}

func NewJQMaster(r *redis.Client) *JQMaster {
	conn := transport.NewConn(r)
	return &JQMaster{
		r:     r,
		conn:  conn,
		sched: scheduler.NewScheduler(r, conn),
	}
}

func (m *JQMaster) distributeJobs(ctx context.Context) error {
	for {
		workerAvailable := true
		if workerAvailable {
			job, err := m.sched.BlockJobPop(ctx)
			if err != nil {
				log.Printf("error getting job: %v", err)
				continue
			}

			if err := m.conn.DistributeJob(ctx, job.ID); err != nil {
				log.Printf("error distributing job: %v", err)
				continue
			}
		}
	}
}

func (m *JQMaster) Run(ctx context.Context) error {
	workerChan := make(chan transport.WorkerRegisterRequest)
	jobChan := make(chan transport.JobRegisterRequest)
	resultChan := make(chan transport.JobResult)

	go m.conn.PollNewClient(ctx, workerChan)
	go m.conn.PollNewJob(ctx, jobChan)
	go m.conn.PollNewResult(ctx, resultChan)

	go m.distributeJobs(ctx)

	for {
		select {
		case newWorker := <-workerChan:
			m.sched.AddWorker(scheduler.NewWorker(newWorker.ID, newWorker.WorkerName, newWorker.Processes))
		case newJob := <-jobChan:
			if err := m.sched.AddJob(ctx, scheduler.Job{
				ID:           newJob.ID,
				Name:         newJob.Name,
				Argument:     newJob.Argument,
				Priority:     newJob.Priority,
				MaxRetry:     newJob.MaxRetry,
				KeepResult:   newJob.KeepResult,
				Timeout:      time.Duration(newJob.Timeout) * time.Second,
				RegisteredAt: time.Now(),
			}); err != nil {
				log.Printf("error adding job: %v", err)
			}
		case newResult := <-resultChan:
			if err := m.sched.ProcessResult(ctx, newResult); err != nil {
				log.Printf("error processing result: %v", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func main() {

}
