package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	cl := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:36381",
	})
	for {
		fmt.Println(cl.Ping().String())
		time.Sleep(time.Millisecond * 200)
	}
}
