//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func TestSentinelConnection(t *testing.T) {
	addrs := requireEnv(t, "REDIS_TUI_TEST_SENTINEL_ADDRS")
	master := requireEnv(t, "REDIS_TUI_TEST_MASTER_NAME")
	profile := redis.ConnectionProfile{
		Name:          "sentinel",
		Addr:          strings.Split(addrs, ",")[0],
		Sentinel:      true,
		SentinelAddrs: strings.Split(addrs, ","),
		MasterName:    master,
	}
	executor := connect(t, profile)

	result, err := executor.Execute(context.Background(), []any{"PING"})
	if err != nil {
		t.Fatalf("PING returned error: %v", err)
	}
	if result.Display() != "PONG" {
		t.Fatalf("PING result = %q, want PONG", result.Display())
	}
}
