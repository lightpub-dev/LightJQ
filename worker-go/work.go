package worker_go

import (
	"fmt"
	"log"
)

func (w *Worker) jqKey() string {
	return fmt.Sprintf("jq:queue:%s", w.workerName)
}

type Job

func (w *Worker) watchForJob() error {
	jqKey := w.jqKey()

	for {
		job, err := w.r.BLPop(w.ctx, 0, jqKey).Result()
		if err != nil {
			log.Fatalf("failed to BLPop: %v", err)
		}

	}
}
