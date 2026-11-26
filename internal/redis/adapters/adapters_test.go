package adapters

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/codebuddy/redis-tui/internal/redis"
)

func newTestManager(t *testing.T) (*Manager, context.Context) {
	t.Helper()
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return New(redis.NewExecutor(client)), context.Background()
}

func TestScanKeysAndDescribeKeys(t *testing.T) {
	manager, ctx := newTestManager(t)
	if err := manager.SetString(ctx, "user:1", "alice", 30*time.Second); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	if err := manager.SetString(ctx, "cache:1", "value", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}

	keys, _, err := manager.ScanKeys(ctx, 0, "user:*", 100)
	if err != nil {
		t.Fatalf("ScanKeys returned error: %v", err)
	}
	if len(keys) != 1 || keys[0] != "user:1" {
		t.Fatalf("keys = %#v, want [user:1]", keys)
	}

	infos, err := manager.DescribeKeys(ctx, keys)
	if err != nil {
		t.Fatalf("DescribeKeys returned error: %v", err)
	}
	if len(infos) != 1 || infos[0].Type != "string" || infos[0].TTL <= 0 {
		t.Fatalf("infos = %#v, want string key with TTL", infos)
	}
}

func TestStringOperations(t *testing.T) {
	manager, ctx := newTestManager(t)
	if err := manager.SetString(ctx, "greeting", "hello", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	value, err := manager.GetString(ctx, "greeting")
	if err != nil {
		t.Fatalf("GetString returned error: %v", err)
	}
	if value != "hello" {
		t.Fatalf("value = %q, want hello", value)
	}
}

func TestHashListSetZSetAndStreamAdapters(t *testing.T) {
	manager, ctx := newTestManager(t)
	client := manager.executor.Client()

	if err := client.HSet(ctx, "user:hash", "name", "alice", "city", "sz").Err(); err != nil {
		t.Fatalf("HSet returned error: %v", err)
	}
	fields, _, err := manager.HScan(ctx, "user:hash", 0, 100)
	if err != nil {
		t.Fatalf("HScan returned error: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("hash fields = %#v, want 2 fields", fields)
	}

	if err := client.RPush(ctx, "queue", "a", "b", "c").Err(); err != nil {
		t.Fatalf("RPush returned error: %v", err)
	}
	items, err := manager.LRange(ctx, "queue", 0, -1)
	if err != nil {
		t.Fatalf("LRange returned error: %v", err)
	}
	if len(items) != 3 || items[2] != "c" {
		t.Fatalf("list items = %#v", items)
	}

	if err := client.SAdd(ctx, "tags", "go", "redis").Err(); err != nil {
		t.Fatalf("SAdd returned error: %v", err)
	}
	members, _, err := manager.SScan(ctx, "tags", 0, 100)
	if err != nil {
		t.Fatalf("SScan returned error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("set members = %#v, want 2", members)
	}

	if err := client.ZAdd(ctx, "ranking", goredis.Z{Score: 10, Member: "alice"}, goredis.Z{Score: 20, Member: "bob"}).Err(); err != nil {
		t.Fatalf("ZAdd returned error: %v", err)
	}
	zmembers, _, err := manager.ZScan(ctx, "ranking", 0, 100)
	if err != nil {
		t.Fatalf("ZScan returned error: %v", err)
	}
	if len(zmembers) != 2 {
		t.Fatalf("zset members = %#v, want 2", zmembers)
	}

	if _, err := client.XAdd(ctx, &goredis.XAddArgs{
		Stream: "events",
		Values: map[string]interface{}{"user": "alice"},
	}).Result(); err != nil {
		t.Fatalf("XAdd returned error: %v", err)
	}
	entries, err := manager.XRange(ctx, "events", "-", "+", 10)
	if err != nil {
		t.Fatalf("XRange returned error: %v", err)
	}
	if len(entries) != 1 || entries[0].Values["user"] != "alice" {
		t.Fatalf("stream entries = %#v", entries)
	}
}

func TestDeleteKeys(t *testing.T) {
	manager, ctx := newTestManager(t)
	if err := manager.SetString(ctx, "gone", "1", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	deleted, err := manager.DeleteKeys(ctx, []string{"gone"})
	if err != nil {
		t.Fatalf("DeleteKeys returned error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}
}
