//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func TestClusterConnection(t *testing.T) {
	addr := requireEnv(t, "REDIS_TUI_TEST_CLUSTER_ADDR")
	executor := connect(t, redis.ConnectionProfile{Name: "cluster", Addr: addr, Cluster: true})

	result, err := executor.Execute(context.Background(), []any{"PING"})
	if err != nil {
		t.Fatalf("PING returned error: %v", err)
	}
	if result.Display() != "PONG" {
		t.Fatalf("PING result = %q, want PONG", result.Display())
	}
}
