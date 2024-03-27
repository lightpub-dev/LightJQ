package scheduler

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Worker struct {
	ID           string
	WorkerName   string
	MaxProcesses int
}

type Job struct {
	ID         string
	Name       string
	Argument   map[string]interface{}
	Priority   int
	MaxRetry   int
	KeepResult bool
	Timeout    time.Duration
}

type Scheduler struct {
	r       *redis.Client
	workers []*Worker

	occupiedWorkers int
	maxProcesses    int
}

func NewWorker(id, workerName string, maxProcesses int) *Worker {
	return &Worker{
		ID:           id,
		WorkerName:   workerName,
		MaxProcesses: maxProcesses,
	}
}

func NewScheduler(r *redis.Client) *Scheduler {
	return &Scheduler{
		r:               r,
		workers:         make([]*Worker, 0),
		occupiedWorkers: 0,
		maxProcesses:    0,
	}
}

func (s *Scheduler) AddWorker(w *Worker) {
	s.workers = append(s.workers, w)
	s.maxProcesses += w.MaxProcesses
}

func (s *Scheduler) HasEmpty() bool {
	return s.occupiedWorkers < s.maxProcesses
}

func (s *Scheduler) AddJob() {

}
