package internal

import (
	"fmt"
	"testing"
)

func TestSetup(t *testing.T, opt ...ClientOption) *Client {
	t.Helper()
	client := NewClient(RedisOpt{
		Addr: "localhost:6379",
		User: "",
		Pass: "",
	}, opt...)
	flushAll(t, client)
	return client
}

func TestTeardown(t *testing.T, client *Client) {
	t.Helper()
	flushAll(t, client)
}

func flushAll(t *testing.T, client *Client) {
	t.Helper()
	err := client.Worker.FlushAll()
	if err != nil {
		fmt.Println(err)
		return
	}
}
