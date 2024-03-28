package transport

import (
	"context"
	"log"

	"github.com/vmihailenco/msgpack/v5"
)

func MakeJobKey(jobID string) string {
	return "jq:job:" + jobID
}

func (c *Conn) DistributeJob(ctx context.Context, jobID string) error {
	_, err := c.r.RPush(ctx, RGlobalQueue, jobID).Result()
	if err != nil {
		return err
	}
	log.Printf("job %s distributed", jobID)
	return nil
}

func (c *Conn) PublishResult(ctx context.Context, result JobResult) error {
	resultBin, err := msgpack.Marshal(&result)
	if err != nil {
		return err
	}

	if _, err := c.r.Publish(ctx, RResultPubSubKey, resultBin).Result(); err != nil {
		return err
	}

	return nil
}
