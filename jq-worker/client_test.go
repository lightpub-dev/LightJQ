package main

import (
	"context"
	"github.com/lightpub-dev/lightjq/jq-worker/internal"
	"testing"
)

func TestWorkerRegister(t *testing.T) {
	t.Helper()
	client := internal.TestSetup(t)

	scenario := struct {
		workerId        string
		workerName      string
		workerProcesses int
	}{
		workerId:        "2",
		workerName:      "worker-1",
		workerProcesses: 1,
	}

	// 1. Register worker
	err := client.Register(context.Background(), &internal.WorkerInfo{
		Id:        scenario.workerId,
		Name:      scenario.workerName,
		Processes: scenario.workerProcesses,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Worker registered successfully")

	// 2. Get worker info
	res := client.Self().LRange(context.Background(), internal.WorkerRegisterQueue, 0, -1)
	if res.Err() != nil {
		t.Fatal(res.Err())
	}

	// 3. Decode worker info
	var workerInfo internal.WorkerInfo
	// check if the list has only one element
	if len(res.Val()) == 0 {
		t.Fatal("worker info not found")
	}
	if len(res.Val()) > 1 {
		t.Fatal("multiple worker info found")
	}
	err = workerInfo.Decode([]byte(res.Val()[0]))
	if err != nil {
		t.Fatal(err)
	}

	// 4. Check worker info
	if workerInfo.Id != scenario.workerId {
		t.Fatalf("expected worker id to be %s, got %s", scenario.workerId, workerInfo.Id)
	}
	if workerInfo.Name != scenario.workerName {
		t.Fatalf("expected worker name to be %s, got %s", scenario.workerName, workerInfo.Name)
	}
	if workerInfo.Processes != scenario.workerProcesses {
		t.Fatalf("expected worker processes to be %d, got %d", scenario.workerProcesses, workerInfo.Processes)
	}

	internal.TestTeardown(t, client)
}
