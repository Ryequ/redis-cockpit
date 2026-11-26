//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/codebuddy/redis-tui/internal/redis"
)

func TestTLSConnection(t *testing.T) {
	addr := requireEnv(t, "REDIS_TUI_TEST_TLS_ADDR")
	profile := redis.ConnectionProfile{
		Name: "tls",
		Addr: addr,
		TLS:  true,
	}
	profile.Password = os.Getenv("REDIS_TUI_TEST_TLS_PASSWORD")
	executor := connect(t, profile)

	result, err := executor.Execute(context.Background(), []any{"PING"})
	if err != nil {
		t.Fatalf("PING returned error: %v", err)
	}
	if result.Display() != "PONG" {
		t.Fatalf("PING result = %q, want PONG", result.Display())
	}
}
