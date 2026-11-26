package redis

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestNewClientForSingleInstance(t *testing.T) {
	server := miniredis.RunT(t)
	profile := ConnectionProfile{Name: "local", Addr: server.Addr()}

	client, err := NewClient(context.Background(), profile)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
}

func TestSentinelAddressesPrefersExplicitList(t *testing.T) {
	profile := ConnectionProfile{
		Addr:          "master.example.com:6379",
		SentinelAddrs: []string{"s1.example.com:26379", "s2.example.com:26379"},
	}
	got := sentinelAddresses(profile)
	want := []string{"s1.example.com:26379", "s2.example.com:26379"}
	if len(got) != len(want) {
		t.Fatalf("sentinel addresses = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sentinel addresses = %#v, want %#v", got, want)
		}
	}
}

func TestSentinelAddressesFallsBackToAddr(t *testing.T) {
	profile := ConnectionProfile{Addr: "127.0.0.1:26379"}
	got := sentinelAddresses(profile)
	if len(got) != 1 || got[0] != "127.0.0.1:26379" {
		t.Fatalf("sentinel addresses = %#v, want [127.0.0.1:26379]", got)
	}
}

func TestNewClientRejectsClusterAndSentinelTogether(t *testing.T) {
	profile := ConnectionProfile{
		Name:       "invalid",
		Addr:       "127.0.0.1:6379",
		Cluster:    true,
		Sentinel:   true,
		MasterName: "mymaster",
	}

	if _, err := NewClient(context.Background(), profile); err == nil {
		t.Fatal("NewClient returned nil error, want validation error")
	}
}
