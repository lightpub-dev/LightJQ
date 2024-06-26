package main

import (
	"context"
	"log"
	"os"
	"strconv"
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

func (m *JQMaster) Run(ctx context.Context) error {
	workerChan := make(chan transport.WorkerRegisterRequest)
	jobChan := make(chan transport.JobRegisterRequest)
	resultChan := make(chan transport.JobResult)

	go m.conn.PollNewClient(ctx, workerChan)
	go m.conn.PollNewJob(ctx, jobChan)
	go m.conn.PollNewResult(ctx, resultChan)

	go m.sched.DistributeJobs(ctx)

	for {
		select {
		case newWorker := <-workerChan:
			m.sched.AddWorker(scheduler.NewWorker(newWorker.ID, newWorker.WorkerName, newWorker.Processes))
			go transport.NewPingChecker(m.conn, newWorker.ID).PingCheckLoop(ctx, func(failure transport.PingFailure) {
				log.Printf("dropping worker %s (ping check failed)", failure.WorkerID)
				m.sched.RemoveWorker(failure.WorkerID)
			})
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
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPort := os.Getenv("REDIS_PORT")
	redisUser := os.Getenv("REDIS_USER")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDatabaseStr := os.Getenv("REDIS_DATABASE")

	if redisAddr == "" {
		redisAddr = "localhost"
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	redisDatabase := 0
	if redisDatabaseStr != "" {
		redisDatabaseInt, err := strconv.Atoi(redisDatabaseStr)
		if err != nil {
			log.Fatalf("invalid REDIS_DATABASE: %v", err)
		}
		redisDatabase = redisDatabaseInt
	}

	r := redis.NewClient(&redis.Options{
		Addr:     redisAddr + ":" + redisPort,
		Username: redisUser,
		Password: redisPassword,
		DB:       redisDatabase,
	})

	master := NewJQMaster(r)
	ctx := context.Background()
	log.Printf("jq-master started")
	if err := master.Run(ctx); err != nil {
		log.Fatalf("error running jq-master: %v", err)
		os.Exit(1)
	}
}
