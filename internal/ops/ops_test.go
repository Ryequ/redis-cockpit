package ops

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func TestParseInfoGroupsSections(t *testing.T) {
	raw := "# Server\r\nredis_version:7.2.0\r\nredis_mode:standalone\r\n\r\n# Memory\r\nused_memory:1024\r\n"
	info := ParseInfo(raw)

	if got := info["Server"]["redis_version"]; got != "7.2.0" {
		t.Fatalf("redis_version = %q, want 7.2.0", got)
	}
	if got := info["Memory"]["used_memory"]; got != "1024" {
		t.Fatalf("used_memory = %q, want 1024", got)
	}
}

func TestParseClientList(t *testing.T) {
	raw := "id=1 addr=127.0.0.1:6000 name=app db=0 cmd=get\nid=2 addr=127.0.0.1:6001 db=2 cmd=set\n"
	clients := ParseClientList(raw)

	if len(clients) != 2 {
		t.Fatalf("clients = %#v, want 2", clients)
	}
	if clients[0].ID != "1" || clients[0].Name != "app" || clients[1].DB != "2" {
		t.Fatalf("clients = %#v", clients)
	}
}

func TestParseConfigReply(t *testing.T) {
	config := ParseConfig([]any{"maxmemory", "0", "appendonly", "yes"})
	if config["maxmemory"] != "0" || config["appendonly"] != "yes" {
		t.Fatalf("config = %#v", config)
	}
}

func TestParseSlowlog(t *testing.T) {
	reply := []any{
		[]any{int64(1), int64(1700000000), int64(25), []any{"SET", "key", "value"}, "127.0.0.1:6000", "app"},
	}
	entries := ParseSlowlog(reply)

	if len(entries) != 1 {
		t.Fatalf("entries = %#v, want 1", entries)
	}
	if entries[0].ID != 1 || entries[0].Duration != 25 || len(entries[0].Command) != 3 {
		t.Fatalf("entry = %#v", entries[0])
	}
}

func TestManagerEvaluatesLuaScript(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	manager := New(redis.NewExecutor(client))
	value, err := manager.Eval(context.Background(), "return redis.call('SET', KEYS[1], ARGV[1])", []string{"lua:key"}, "lua-value")
	if err != nil {
		t.Fatalf("Eval returned error: %v", err)
	}
	if value.Display() != "OK" {
		t.Fatalf("Eval result = %q, want OK", value.Display())
	}
}

func TestManagerRunsTransaction(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	manager := New(redis.NewExecutor(client))
	values, err := manager.Transaction(context.Background(), [][]any{
		{"SET", "tx:key", "1"},
		{"GET", "tx:key"},
	})
	if err != nil {
		t.Fatalf("Transaction returned error: %v", err)
	}
	if len(values) != 2 || values[0].Display() != "OK" || values[1].Display() != "1" {
		t.Fatalf("transaction results = %#v", values)
	}
}

func TestManagerSnapshotUsesRedisCommands(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	manager := New(redis.NewExecutor(client))
	snapshot, err := manager.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if len(snapshot.Info) == 0 {
		t.Fatal("snapshot.Info is empty")
	}
}
