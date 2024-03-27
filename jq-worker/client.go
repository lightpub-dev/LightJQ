package main

import (
	"context"
	"github.com/lightpub-dev/lightjq/jq-worker/internal"
)

func main() {
	client := internal.NewClient(internal.RedisOpt{
		Addr: "localhost:6379",
		User: "",
		Pass: "",
	})

	err := client.Register(context.Background(), &internal.WorkerInfo{
		Id:        "1",
		Name:      "worker-1",
		Processes: 1,
	})
	if err != nil {
		return
	}
}
