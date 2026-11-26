package stream

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func TestBoundedDropsOldestMessages(t *testing.T) {
	bounded := NewBounded[int](2)
	for i := 1; i <= 5; i++ {
		bounded.Send(context.Background(), i)
	}

	got := make([]int, 0, 2)
	for len(bounded.C()) > 0 {
		got = append(got, <-bounded.C())
	}
	if len(got) != 2 || got[0] != 4 || got[1] != 5 {
		t.Fatalf("messages = %#v, want [4 5]", got)
	}
	if bounded.Dropped() != 3 {
		t.Fatalf("dropped = %d, want 3", bounded.Dropped())
	}
}

func TestSubscriptionCloseSignalsDone(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	hub := NewHub(client)
	subscription, err := hub.Subscribe(context.Background(), "done-channel")
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	if err := subscription.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	select {
	case <-subscription.Done():
	case <-time.After(time.Second):
		t.Fatal("Done was not signaled after Close")
	}
}

func TestSubscriptionSurvivesParentContextCancellation(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	hub := NewHub(client)
	parent, cancelParent := context.WithCancel(context.Background())
	subscription, err := hub.Subscribe(parent, "survive")
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	defer subscription.Close()
	cancelParent()

	if err := hub.Publish(context.Background(), "survive", "still-alive"); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	select {
	case message := <-subscription.Messages():
		if message.Payload != "still-alive" {
			t.Fatalf("payload = %q, want still-alive", message.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("subscription stopped after parent context cancellation")
	}
}

func TestPubSubHubReceivesPublishedMessage(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	hub := NewHub(client)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	subscription, err := hub.Subscribe(ctx, "news")
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	defer subscription.Close()

	if err := hub.Publish(ctx, "news", "hello"); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	select {
	case message := <-subscription.Messages():
		if message.Channel != "news" || message.Payload != "hello" {
			t.Fatalf("message = %#v", message)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for pub/sub message")
	}
}
