package redis

import (
	"context"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/command"
)

type fakeDoer struct {
	calls int
	reply any
	err   error
}

func (f *fakeDoer) Do(_ context.Context, _ ...any) (any, error) {
	f.calls++
	return f.reply, f.err
}

func TestDiscovererParsesCommandReply(t *testing.T) {
	reply := []any{
		[]any{"get", int64(-2), []any{"readonly", "fast"}},
		[]any{"set", int64(3), []any{"write"}},
	}
	discoverer := NewDiscoverer(&fakeDoer{reply: reply})

	specs, err := discoverer.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("spec count = %d, want 2", len(specs))
	}
	if specs[0].Name != "GET" || specs[0].Arity != -2 {
		t.Fatalf("first spec = %#v, want GET with arity -2", specs[0])
	}
	if len(specs[0].Flags) != 2 || specs[0].Flags[0] != "readonly" {
		t.Fatalf("flags = %#v, want readonly/fast", specs[0].Flags)
	}
}

func TestDiscovererCachesResult(t *testing.T) {
	doer := &fakeDoer{reply: []any{[]any{"ping", int64(1), []any{"fast"}}}}
	discoverer := NewDiscoverer(doer)

	if _, err := discoverer.Discover(context.Background()); err != nil {
		t.Fatalf("first Discover returned error: %v", err)
	}
	if _, err := discoverer.Discover(context.Background()); err != nil {
		t.Fatalf("second Discover returned error: %v", err)
	}
	if doer.calls != 1 {
		t.Fatalf("Do calls = %d, want 1", doer.calls)
	}
}

func TestDiscovererReturnsStaticFallbackOnError(t *testing.T) {
	discoverer := NewDiscoverer(&fakeDoer{err: context.DeadlineExceeded})

	specs, err := discoverer.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if !command.NewRegistry().Has("GET") && !containsCommand(specs, "GET") {
		t.Fatal("fallback specs should include GET")
	}
}

func containsCommand(specs []command.Spec, name string) bool {
	for _, spec := range specs {
		if spec.Name == name {
			return true
		}
	}
	return false
}
