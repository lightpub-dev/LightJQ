package internal

import (
	"fmt"
	"testing"
)

func TestSetup(t *testing.T) *Client {
	t.Helper()
	client := NewClient(RedisOpt{
		Addr: "localhost:6379",
		User: "",
		Pass: "",
	})
	flushAll(t, client)
	return client
}

func flushAll(t *testing.T, client *Client) {
	t.Helper()
	err := client.FlushAll()
	if err != nil {
		fmt.Println(err)
		return
	}
}
