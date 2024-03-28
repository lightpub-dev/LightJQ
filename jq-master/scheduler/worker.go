package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lightpub-dev/lightjq/jq-master/transport"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

func makeJobKey(jobID string) string {
	return transport.MakeJobKey(jobID)
}

const (
	RScoredJobSet   = "jq:scoredJobSet"
	RProcessingJobs = "jq:processingJobs"
)

type Worker struct {
	ID           string
	WorkerName   string
	MaxProcesses int
}

type Job struct {
	ID           string                 `msgpack:"id"`
	Name         string                 `msgpack:"name"`
	Argument     map[string]interface{} `msgpack:"argument"`
	Priority     int                    `msgpack:"priority"`
	MaxRetry     int                    `msgpack:"max_retry"`
	CurrentRetry int                    `msgpack:"current_retry"`
	KeepResult   bool                   `msgpack:"keep_result"`
	Timeout      time.Duration          `msgpack:"timeout"`
	RegisteredAt time.Time              `msgpack:"registered_at"`
}

func (j Job) CalculatePriorityScore() float64 {
	return float64(j.Priority)
}

type Scheduler struct {
	r       *redis.Client
	workers []*Worker

	maxProcesses int

	tran *transport.Conn
}

func NewWorker(id, workerName string, maxProcesses int) *Worker {
	return &Worker{
		ID:           id,
		WorkerName:   workerName,
		MaxProcesses: maxProcesses,
	}
}

func NewScheduler(r *redis.Client, tran *transport.Conn) *Scheduler {
	return &Scheduler{
		r:            r,
		workers:      make([]*Worker, 0),
		maxProcesses: 0,
		tran:         tran,
	}
}

func (s *Scheduler) AddWorker(w *Worker) {
	s.workers = append(s.workers, w)
	s.maxProcesses += w.MaxProcesses
}

func (s *Scheduler) getCurrentProcessingJobsN(ctx context.Context) (int64, error) {
	return s.r.SCard(ctx, RProcessingJobs).Result()
}

func (s *Scheduler) addToProcessingJobs(ctx context.Context, jobID string) error {
	return s.r.SAdd(ctx, RProcessingJobs, jobID).Err()
}

func (s *Scheduler) removeFromProcessingJobs(ctx context.Context, jobID string) error {
	return s.r.SRem(ctx, RProcessingJobs, jobID).Err()
}

func (s *Scheduler) HasEmpty(ctx context.Context) (bool, error) {
	n, err := s.getCurrentProcessingJobsN(ctx)
	if err != nil {
		return false, err
	}
	return n < int64(s.maxProcesses), nil
}

func (s *Scheduler) AddJob(ctx context.Context, job Job) error {
	// priority

	// transaction
	tx := s.r.TxPipeline()

	// register job info to job list
	jobBin, err := msgpack.Marshal(&job)
	if err != nil {
		return err
	}
	if _, err := tx.Set(ctx, makeJobKey(job.ID), jobBin, 0).Result(); err != nil {
		return err
	}

	// push job id to scored job set
	if _, err := tx.ZAdd(ctx, RScoredJobSet, redis.Z{
		Score:  job.CalculatePriorityScore(),
		Member: job.ID,
	}).Result(); err != nil {
		return err
	}

	// commit
	if _, err := tx.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) BlockJobPop(ctx context.Context) (Job, error) {
	// pop job from scored job set
	js, err := s.r.BZPopMin(ctx, 0, RScoredJobSet).Result()
	if err != nil {
		return Job{}, err
	}

	// get job info
	jobBin, err := s.r.Get(ctx, makeJobKey(js.Member.(string))).Bytes()
	if err != nil {
		return Job{}, err
	}

	var job Job
	if err := msgpack.Unmarshal(jobBin, &job); err != nil {
		return Job{}, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return job, nil
}

func (s *Scheduler) DistributeJobs(ctx context.Context) error {
	for {
		workerAvailable, err := s.HasEmpty(ctx)
		if err != nil {
			log.Printf("failed to scard processing jobs: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if workerAvailable {
			job, err := s.BlockJobPop(ctx)
			if err != nil {
				log.Printf("error getting job: %v", err)
				continue
			}

			if err := s.addToProcessingJobs(ctx, job.ID); err != nil {
				log.Printf("error adding job to processing jobs: %v", err)
				log.Printf("ignoring this error and continue")
			}

			if err := s.tran.DistributeJob(ctx, job.ID); err != nil {
				log.Printf("error distributing job: %v", err)
				if err := s.removeFromProcessingJobs(ctx, job.ID); err != nil {
					log.Printf("error removing job from processing jobs: %v", err)
				}
				continue
			}
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
