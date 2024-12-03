package redisprefixtest

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRaceCondition(t *testing.T) {
	t.Parallel()
	m := miniredis.RunT(t)

	t.Run("balance-150", func(t *testing.T) {
		t.Parallel()
		client := newClusterClient(m.Addr())
		client.AddHook(New(t))
		setAndGet(t, client, "user:1:balance", "150")
	})

	t.Run("balance-200", func(t *testing.T) {
		t.Parallel()
		client := newClusterClient(m.Addr())
		client.AddHook(New(t))
		setAndGet(t, client, "user:1:balance", "200")
	})
}

func TestDataRacePipeline(t *testing.T) {
	t.Parallel()
	m := miniredis.RunT(t)

	for range 3 {
		t.Run("incr", func(t *testing.T) {
			t.Parallel()
			client := newClusterClient(m.Addr())
			client.AddHook(New(t))

			pipe := client.Pipeline()
			pipe.Incr(context.TODO(), "X")
			pipe.Incr(context.TODO(), "X")
			pipe.Incr(context.TODO(), "X")
			result, err := pipe.Exec(context.TODO())
			if err != nil {
				t.Fatal(err)
			}
			for i, cmd := range result {
				got, err := cmd.(*redis.IntCmd).Result()
				if err != nil {
					t.Fatal(err)
				}
				if int(got) != i+1 {
					t.Errorf("expected %d, got: %d", i+1, got)
				}
			}
		})
	}
}

func TestCommandUnchanged(t *testing.T) {
	t.Parallel()
	m := miniredis.RunT(t)
	client := newClusterClient(m.Addr())
	client.AddHook(New(t))

	cmd := client.Set(context.TODO(), "foo", "bar", 0)
	if got, want := cmd.String(), "set foo bar: OK"; got != want {
		t.Errorf("got %s, want: %s", got, want)
	}
}

func TestWithDelimiter(t *testing.T) {
	t.Parallel()
	m := miniredis.RunT(t)
	client := newClusterClient(m.Addr())
	client.AddHook(New(t, WithDelimiter("|")))

	if err := client.Set(context.TODO(), "user:1:balance", "150", 0).Err(); err != nil {
		t.Fatal(err)
	}

	clientWithoutHook := newClusterClient(m.Addr())
	if err := clientWithoutHook.Get(context.TODO(), "TestWithDelimiter|user:1:balance").Err(); err != nil {
		t.Fatal(err)
	}
}

func TestWithDelimiter_Empty(t *testing.T) {
	t.Parallel()
	var options options
	err := WithOptions(WithDelimiter(""))(&options)
	if err == nil {
		t.Fatal("expected an error")
	}
	if got, want := err.Error(), "WithDelimiter: cannot be empty"; got != want {
		t.Errorf("got %s, want: %s", got, want)
	}
}

func TestInvalidKey(t *testing.T) {
	t.Parallel()
	m := miniredis.RunT(t)
	client := newClusterClient(m.Addr())
	client.AddHook(New(t))

	if _, err := client.Do(context.TODO(), "SELECT", 2).Result(); err != nil {
		t.Fatal(err)
	}
}

func TestNoArgsCommand(t *testing.T) {
	t.Parallel()
	m := miniredis.RunT(t)
	client := newClusterClient(m.Addr())
	client.AddHook(New(t))

	if _, err := client.Command(context.TODO()).Result(); err != nil {
		t.Fatal(err)
	}
}

func setAndGet(tb testing.TB, client *redis.ClusterClient, key, value string) {
	tb.Helper()
	resp, err := client.Set(context.TODO(), key, value, 0).Result()
	if err != nil {
		tb.Fatal(err)
	}
	if got, want := resp, "OK"; got != want {
		tb.Errorf("got %s, want: %s", got, want)
	}
	resp, err = client.Get(context.TODO(), key).Result()
	if err != nil {
		tb.Fatal(err)
	}
	if got, want := resp, value; got != want {
		tb.Errorf("got %s, want: %s", got, want)
	}
}

func newClusterClient(addr string) *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			addr,
		},
	})
}
