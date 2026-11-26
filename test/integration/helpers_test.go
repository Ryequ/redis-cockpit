//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/codebuddy/redis-tui/internal/redis"
)

func requireEnv(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Skipf("%s is not set", name)
	}
	return value
}

func connect(t *testing.T, profile redis.ConnectionProfile) *redis.Executor {
	t.Helper()
	client, err := redis.NewClient(context.Background(), profile)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return redis.NewExecutor(client)
}
