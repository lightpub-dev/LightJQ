package main

import (
	"context"
	"github.com/lightpub-dev/lightjq/jq-worker/internal"
	"testing"
)

func TestWorkerRegister(t *testing.T) {
	t.Helper()
	client := internal.TestSetup(t)
	defer client.Worker.Close()

	// 1. Register worker
	err := client.Register(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Worker registered successfully")
	expectedWorkerInfo := client.Info

	// 2. Get worker info
	redisClient := client.Worker.(internal.RedisConn).Client
	res := redisClient.LRange(context.Background(), internal.WorkerRegisterQueue, 0, -1)
	if res.Err() != nil {
		t.Fatal(res.Err())
	}

	// 3. Decode worker info
	var decWorkerInfo internal.WorkerInfo
	// check if the list has only one element
	if len(res.Val()) == 0 {
		t.Fatal("worker info not found")
	}
	if len(res.Val()) > 1 {
		t.Fatal("multiple worker info found")
	}
	err = decWorkerInfo.Decode([]byte(res.Val()[0]))
	if err != nil {
		t.Fatal(err)
	}

	// 4. Check worker info
	if decWorkerInfo.Id != expectedWorkerInfo.Id {
		t.Fatalf("expected worker id to be %s, got %s", expectedWorkerInfo.Id, decWorkerInfo.Id)
	}
	if decWorkerInfo.Name != expectedWorkerInfo.Name {
		t.Fatalf("expected worker name to be %s, got %s", expectedWorkerInfo.Name, decWorkerInfo.Name)
	}
	if decWorkerInfo.Processes != expectedWorkerInfo.Processes {
		t.Fatalf("expected worker processes to be %d, got %d", expectedWorkerInfo.Processes, decWorkerInfo.Processes)
	}

	internal.TestTeardown(t, client)
}
