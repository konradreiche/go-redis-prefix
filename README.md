# 🏷️ go-redis-prefix
[![ci](https://github.com/konradreiche/go-redis-prefix/actions/workflows/ci.yaml/badge.svg)](https://github.com/konradreiche/go-redis-prefix/actions) [![codecov](https://codecov.io/gh/konradreiche/go-redis-prefix/graph/badge.svg?token=kXoAXWhLJS)](https://codecov.io/gh/konradreiche/go-redis-prefix)

A package that provides a Redis hook to automatically prefix keys during testing. This helps prevent key collisions when running tests in parallel or across multiple packages that interact with the same Redis instance.

## Installation

```
go get github.com/konradreiche/go-redis-prefix@latest
```

## Usage

```go
import (
	"testing"

	"github.com/konradreiche/go-redis-prefix/redisprefixtest"
	"github.com/redis/go-redis/v9"
)

func TestRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client.AddHook(redisprefixtest.New(t))

    // Your test code here...
}
```

## Limitations

* **Multi-Key Commands**: The hook does not support automatic prefixing for multi-key commands. You must handle these cases manually or avoid using multi-key commands in tests.
* **Key Position**: The hook assumes that the key is the command's second argument. This is true for most commands, but some have keys in different positions.
