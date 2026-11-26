package command

import (
	"reflect"
	"testing"
)

func TestParseSplitsQuotedArguments(t *testing.T) {
	got, err := Parse(`SET "greeting text" 'hello world' 60`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	want := []string{"SET", "greeting text", "hello world", "60"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestParseRejectsUnclosedQuote(t *testing.T) {
	if _, err := Parse(`SET "unclosed`); err == nil {
		t.Fatal("Parse returned nil error for unclosed quote")
	}
}

func TestParseReturnsEmptyArgsForEmptyLine(t *testing.T) {
	got, err := Parse("   ")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("args = %#v, want empty", got)
	}
}
