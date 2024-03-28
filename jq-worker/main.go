package main

import (
	"context"
	"fmt"
	"github.com/lightpub-dev/lightjq/jq-worker/internal"
	"time"
)

func main() {
	processes := 5

	client := internal.NewClient(internal.RedisOpt{
		Addr: "localhost:6379",
		User: "",
		Pass: "",
	}, internal.WithProcesses(processes), internal.WithHostname("worker-1"))
	defer client.Close()

	err := client.Register(context.Background())
	if err != nil {
		return
	}

	// Example of how to enqueue jobs
	for i := 0; i < 10; i++ {
		err = client.Enqueue(context.Background(), &internal.JobInfo{
			Id:   fmt.Sprintf("job-%d", i),
			Name: fmt.Sprintf("job-%d", i),
			Argument: map[string]interface{}{
				"key":  "value",
				"key2": "value2",
			},
			Priority: 10,
			MaxRetry: 1,
		})
		if err != nil {
			return
		}
	}

	// make sure to run this in a separate goroutine
	waitCh := make(chan struct{})
	parallelCh := make(chan struct{}, processes)
	go func() {
		for {
			// limit the number of parallel processes
			parallelCh <- struct{}{}
			go func() {
				defer func() {
					<-parallelCh
				}()
				for {
					job, err := getJob(client)
					if err != nil {
						continue
					}
					if job == nil {
						time.Sleep(1 * time.Second)
						continue
					}
					err = doProcess(job)
					if err != nil {
						fmt.Println(err)
					}
				}
			}()
		}
	}()

	<-waitCh

}

func getJob(client *internal.Client) (*internal.JobInfo, error) {
	job, err := client.Dequeue(context.Background())
	if err != nil {
		return nil, err
	}
	return job, nil
}

func doProcess(job *internal.JobInfo) error {
	fmt.Printf("Processing job: %s\n", job.Name)
	time.Sleep(2 * time.Second)
	fmt.Printf("Processed job: %s\n", job.Name)
	return nil
}
