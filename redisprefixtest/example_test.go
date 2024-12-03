package redisprefixtest_test

import (
	"testing"

	"github.com/konradreiche/go-redis-prefix/redisprefixtest"
	"github.com/redis/go-redis/v9"
)

func ExampleNew() {
	var t *testing.T
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client.AddHook(redisprefixtest.New(t))
}
