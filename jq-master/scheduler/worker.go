package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/lightpub-dev/lightjq/jq-master/transport"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

func makeJobKey(jobID string) string {
	return transport.MakeJobKey(jobID)
}

const (
	RScoredJobSet = "jq:scoredJobSet"
)

type Worker struct {
	ID           string
	WorkerName   string
	MaxProcesses int
}

type Job struct {
	ID           string
	Name         string
	Argument     map[string]interface{}
	Priority     int
	MaxRetry     int
	CurrentRetry int
	KeepResult   bool
	Timeout      time.Duration
	RegisteredAt time.Time
}

func (j Job) CalculatePriorityScore() float64 {
	return float64(j.Priority)
}

type Scheduler struct {
	r       *redis.Client
	workers []*Worker

	occupiedWorkers int
	maxProcesses    int

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
		r:               r,
		workers:         make([]*Worker, 0),
		occupiedWorkers: 0,
		maxProcesses:    0,
		tran:            tran,
	}
}

func (s *Scheduler) AddWorker(w *Worker) {
	s.workers = append(s.workers, w)
	s.maxProcesses += w.MaxProcesses
}

func (s *Scheduler) HasEmpty() bool {
	return s.occupiedWorkers < s.maxProcesses
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
