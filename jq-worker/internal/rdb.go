package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOpt is a struct that contains the address, username, and password of a Redis instance.
type RedisOpt struct {
	Addr string
	User string
	Pass string
}

// RedisConn is a struct that holds a connection to a Redis instance
//
// implements the Worker interface
type RedisConn struct {
	Client *redis.Client
}

func (r RedisConn) Close() error {
	return r.Client.Close()
}

func (r RedisConn) Ping(ctx context.Context, workerID string) error {
	pingMsg := PingMessage{
		WorkerID: workerID,
	}
	encMsg, err := pingMsg.Encode()
	if err != nil {
		return err
	}

	// publish the ping message to the worker queue
	return r.Client.Publish(ctx, PingChannel, encMsg).Err()
}

func (r RedisConn) Register(ctx context.Context, info *WorkerInfo) error {
	encMsg, err := info.Encode()
	if err != nil {
		return err
	}
	return r.Client.RPush(ctx, WorkerRegisterQueue, encMsg).Err()
}

// Enqueue **THIS IS A DEBUGGING FUNCTION**
func (r RedisConn) Enqueue(ctx context.Context, job *JobInfo) error {
	encMsg, err := job.Encode()
	if err != nil {
		return err
	}
	return r.Client.RPush(ctx, JobRegisterQueue, encMsg).Err()
}

func (r RedisConn) Dequeue(ctx context.Context) (*JobInfo, error) {
	startedAt := time.Now().Format(time.RFC3339)

	// 1. Pop a job from the global queue
	encMsg, err := r.Client.BLPop(ctx, 0, GlobalQueue).Result()
	if err != nil {
		return nil, err
	}

	// 2. Get the job from the job queue
	jobId := JobQueuePrefix + encMsg[1]
	jobEnc, err := r.Client.Get(ctx, jobId).Result()
	if err != nil {
		return nil, err
	}

	// 3. Decode the job
	var job JobInfo
	if err := job.Decode([]byte(jobEnc)); err != nil {
		return nil, err
	}

	// 4. Notify that the job is being processed
	job.StartedAt = startedAt
	err = r.notifyProcessing(ctx, &job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// notifyProcessing notifies the master that the worker is processing a job
func (r RedisConn) notifyProcessing(ctx context.Context, job *JobInfo) error {
	processingInfo := job.GenerateProcessingInfo()
	encMsg, err := encodeMsg(processingInfo)
	if err != nil {
		return err
	}

	return r.Client.SAdd(ctx, ProcessingSet, encMsg).Err()
}

// removeProcessing_slow removes the job from the processing set
func (r RedisConn) removeProcessing_slow(ctx context.Context, jobId string) error {
	// find the job in the ProcessingSet and remove it
	encMsgs, err := r.Client.SMembers(ctx, ProcessingSet).Result()
	if err != nil {
		return err
	}

	// iterate through the messages and remove the job if found
	// TODO: use a more efficient way to remove the job
	for _, msg := range encMsgs {
		encMsg := []byte(msg)
		var decMsg ProcessingInfo
		err := decodeMsg(encMsg, &decMsg)
		if err != nil {
			return err
		}

		// remove the job if found
		if decMsg.JobID == jobId {
			return r.Client.SRem(ctx, ProcessingSet, encMsg).Err()
		}
	}

	// return an error if the job is not found
	return fmt.Errorf("job not found in the processing set")
}

func (r RedisConn) ReportResult(ctx context.Context, result *JobResult) error {
	// 1. Encode the result
	encMsg, err := encodeMsg(&result)
	if err != nil {
		return err
	}

	// 2. Remove the job from the processing set
	err = r.removeProcessing_slow(ctx, result.JobID)
	if err != nil {
		return err
	}

	// 3. Push the result to the result queue
	return r.Client.RPush(ctx, ResultQueue, encMsg).Err()
}

func (r RedisConn) FlushAll() error {
	return r.Client.FlushAll(context.Background()).Err()
}
