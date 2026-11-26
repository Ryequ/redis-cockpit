//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/stream"
)

func TestPubSubAndStreamCommands(t *testing.T) {
	addr := requireEnv(t, "REDIS_TUI_TEST_ADDR")
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "integration", Addr: addr})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hub := stream.NewHub(client)
	subscription, err := hub.Subscribe(ctx, "integration:channel")
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	defer subscription.Close()

	if err := hub.Publish(ctx, "integration:channel", "hello"); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	select {
	case message := <-subscription.Messages():
		if message.Payload != "hello" {
			t.Fatalf("payload = %q, want hello", message.Payload)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}

	executor := redis.NewExecutor(client)
	if _, err := executor.Execute(ctx, []any{"XADD", "integration:stream", "*", "field", "value"}); err != nil {
		t.Fatalf("XADD returned error: %v", err)
	}
	result, err := executor.Execute(ctx, []any{"XRANGE", "integration:stream", "-", "+"})
	if err != nil {
		t.Fatalf("XRANGE returned error: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("XRANGE returned no entries")
	}
}
