//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/codebuddy/redis-tui/internal/redis"
)

func TestSingleInstanceSetGet(t *testing.T) {
	addr := requireEnv(t, "REDIS_TUI_TEST_ADDR")
	executor := connect(t, redis.ConnectionProfile{Name: "integration", Addr: addr})

	set, err := executor.Execute(context.Background(), []any{"SET", "integration:single", "ok"})
	if err != nil {
		t.Fatalf("SET returned error: %v", err)
	}
	if set.Display() != "OK" {
		t.Fatalf("SET result = %q, want OK", set.Display())
	}

	get, err := executor.Execute(context.Background(), []any{"GET", "integration:single"})
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	if get.Display() != "ok" {
		t.Fatalf("GET result = %q, want ok", get.Display())
	}
}
