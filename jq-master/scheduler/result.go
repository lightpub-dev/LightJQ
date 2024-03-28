package scheduler

import (
	"context"
	"fmt"

	"github.com/lightpub-dev/lightjq/jq-master/transport"
	"github.com/vmihailenco/msgpack/v5"
)

func (s *Scheduler) ProcessResult(ctx context.Context, result transport.JobResult) error {
	switch result.Type {
	case transport.JobResultSuccess:
		// remove job data
		jobKey := makeJobKey(result.JobID)
		if err := s.r.Del(ctx, jobKey).Err(); err != nil {
			return err
		}
		// one worker is now available
		s.removeFromProcessingJobs(ctx, result.JobID)
		// send back the result to pusher
		return s.tran.PublishResult(ctx, result)
	case transport.JobResultFailure:
		return s.retryJob(ctx, result)
	default:
		return fmt.Errorf("unknown job result type: %s", result.Type)
	}
}

func (s *Scheduler) retryJob(ctx context.Context, result transport.JobResult) error {
	jobKey := makeJobKey(result.JobID)
	jobBin, err := s.r.Get(ctx, jobKey).Bytes()
	if err != nil {
		return err
	}

	var job Job
	if err := msgpack.Unmarshal(jobBin, &job); err != nil {
		return err
	}

	if job.MaxRetry == job.CurrentRetry {
		// no more retry left

		// remove job data
		if err := s.r.Del(ctx, jobKey).Err(); err != nil {
			return err
		}
		// one worker is now available
		s.removeFromProcessingJobs(ctx, result.JobID)
		// report error to pusher
		return s.tran.PublishResult(ctx, result)
	}

	job.CurrentRetry++

	// re-enqueue the job
	if err := s.AddJob(ctx, job); err != nil {
		return err
	}

	return nil
}
