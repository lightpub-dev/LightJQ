package main

import "github.com/redis/go-redis/v9"

type JQMaster struct {
	r *redis.Client
}

func NewJQMaster(r *redis.Client) *JQMaster {
	return &JQMaster{r: r}
}

func main() {

}
