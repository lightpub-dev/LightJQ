package main

import (
	"context"
	"fmt"
	"github.com/lightpub-dev/lightjq/jq-worker/internal"
)

func main() {
	client := internal.NewClient(internal.RedisOpt{
		Addr: "localhost:6379",
		User: "",
		Pass: "",
	}, internal.WithProcesses(20), internal.WithHostname("worker-1"))
	defer client.Close()

	err := client.Register(context.Background())
	if err != nil {
		return
	}

	sampleJob := internal.JobInfo{
		Id:   "job-1",
		Name: "job-1",
		Argument: map[string]interface{}{
			"key":  "value",
			"key2": "value2",
		},
		Priority: 10,
		MaxRetry: 1,
	}

	err = client.Enqueue(context.Background(), &sampleJob)
	if err != nil {
		return
	}

	job, err := client.Dequeue(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	// print job info
	fmt.Printf("Job ID: %s\n", job.Id)
	fmt.Printf("Job Name: %s\n", job.Name)
	fmt.Printf("Job Argument: %v\n", job.Argument)
	fmt.Printf("Job Priority: %d\n", job.Priority)
	fmt.Printf("Job MaxRetry: %d\n", job.MaxRetry)

}
