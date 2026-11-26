package redis

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestExecutorExecutesRawCommands(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := NewClient(context.Background(), ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	executor := NewExecutor(client)
	ctx := context.Background()

	setResult, err := executor.Execute(ctx, []any{"SET", "greeting", "hello"})
	if err != nil {
		t.Fatalf("SET returned error: %v", err)
	}
	if setResult.Display() != "OK" {
		t.Fatalf("SET result = %q, want OK", setResult.Display())
	}

	getResult, err := executor.Execute(ctx, []any{"GET", "greeting"})
	if err != nil {
		t.Fatalf("GET returned error: %v", err)
	}
	if getResult.Display() != "hello" {
		t.Fatalf("GET result = %q, want hello", getResult.Display())
	}
}

func TestExecutorRejectsEmptyCommand(t *testing.T) {
	executor := NewExecutor(nil)
	if _, err := executor.Execute(context.Background(), nil); err == nil {
		t.Fatal("Execute returned nil error for empty argv")
	}
}

func TestExecutorExposesUnderlyingClient(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := NewClient(context.Background(), ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	executor := NewExecutor(client)
	if executor.Client() != client {
		t.Fatal("Client did not return the underlying UniversalClient")
	}
}
