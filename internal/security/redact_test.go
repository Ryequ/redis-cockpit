package security

import (
	"reflect"
	"testing"
)

func TestRedactCommandMasksSensitiveArguments(t *testing.T) {
	got := RedactCommand([]string{"AUTH", "default", "secret"})
	want := []string{"AUTH", "default", "***"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("redacted = %#v, want %#v", got, want)
	}

	got = RedactCommand([]string{"CONFIG", "SET", "requirepass", "secret"})
	if got[len(got)-1] != "***" {
		t.Fatalf("redacted = %#v, want last arg masked", got)
	}
}

func TestRedactCommandMasksACLSetUserPasswordTokens(t *testing.T) {
	got := RedactCommand([]string{"ACL", "SETUSER", "alice", ">secret", "on", "~*", "+get"})
	if got[3] != "***" {
		t.Fatalf("password token not masked: %#v", got)
	}
	if got[4] != "on" || got[5] != "~*" {
		t.Fatalf("ACL rules were corrupted: %#v", got)
	}
}

func TestRedactCommandMasksMasterAuth(t *testing.T) {
	got := RedactCommand([]string{"CONFIG", "SET", "masterauth", "secret"})
	if got[3] != "***" {
		t.Fatalf("masterauth was not masked: %#v", got)
	}
}

func TestRedactCommandMasksHelloAuth(t *testing.T) {
	got := RedactCommand([]string{"HELLO", "3", "AUTH", "default", "secret"})
	if got[len(got)-1] != "***" {
		t.Fatalf("redacted = %#v, want last arg masked", got)
	}
}

func TestRedactTextMasksPasswordAssignments(t *testing.T) {
	got := RedactText("addr=127.0.0.1:6379 password=secret db=0")
	if got != "addr=127.0.0.1:6379 password=*** db=0" {
		t.Fatalf("redacted = %q", got)
	}
}
