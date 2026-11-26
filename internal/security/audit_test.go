package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditLogRedactsSensitiveArguments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	audit := NewAuditLog(path)

	if err := audit.Record("local", []string{"CONFIG", "SET", "requirepass", "secret"}); err != nil {
		t.Fatalf("Record returned error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit file: %v", err)
	}
	text := string(data)
	if strings.Contains(text, "secret") {
		t.Fatalf("audit log leaked secret: %s", text)
	}
	if !strings.Contains(text, "***") {
		t.Fatalf("audit log did not mask secret: %s", text)
	}
}
