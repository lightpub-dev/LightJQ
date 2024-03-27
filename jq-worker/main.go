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
	defer client.Close()

	err := client.Register(context.Background())
	if err != nil {
		return
	}
}
