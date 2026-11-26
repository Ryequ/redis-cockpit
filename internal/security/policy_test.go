package security

import "testing"

func TestPolicyBlocksWritesInReadOnlyMode(t *testing.T) {
	p := NewPolicy(true)
	if err := p.Check([]string{"GET", "key"}); err != nil {
		t.Fatalf("GET blocked in read-only mode: %v", err)
	}
	if err := p.Check([]string{"SET", "key", "value"}); err == nil {
		t.Fatal("SET should be blocked in read-only mode")
	}
}

func TestPolicyAllowsCommonReadCommands(t *testing.T) {
	p := NewPolicy(true)
	commands := [][]string{
		{"XREAD", "STREAMS", "events", "0"},
		{"XINFO", "STREAM", "events"},
		{"XPENDING", "events", "group"},
		{"HKEYS", "hash"},
		{"ZRANK", "zset", "member"},
	}
	for _, argv := range commands {
		if err := p.Check(argv); err != nil {
			t.Fatalf("read command %v was blocked: %v", argv, err)
		}
	}
}

func TestPolicyFailsClosedForUnknownAndWriteCommands(t *testing.T) {
	p := NewPolicy(true)
	commands := [][]string{
		{"UNKNOWN"},
		{"RESTORE", "key", "payload"},
		{"SETNX", "key", "value"},
		{"SCRIPT", "FLUSH"},
		{"EVAL", "return 1"},
	}
	for _, argv := range commands {
		if err := p.Check(argv); err == nil {
			t.Fatalf("command %v should be blocked in read-only mode", argv)
		}
	}
}

func TestPolicyRequiresConfirmationForCriticalCommands(t *testing.T) {
	p := NewPolicy(false)
	if !p.RequiresConfirmation([]string{"FLUSHALL"}) {
		t.Fatal("FLUSHALL should require confirmation")
	}
	if p.RequiresConfirmation([]string{"GET", "key"}) {
		t.Fatal("GET should not require confirmation")
	}
}
