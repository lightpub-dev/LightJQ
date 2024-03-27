package main

import (
	"context"
	"github.com/lightpub-dev/lightjq/jq-worker/internal"
	"testing"
)

func TestWorkerRegister(t *testing.T) {
	t.Helper()
	client := internal.TestSetup(t)
	err := client.Register(context.Background(), &internal.WorkerInfo{
		Id:        "2",
		Name:      "worker-1",
		Processes: 1,
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Log("Worker registered successfully")
}
