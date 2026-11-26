package cmd

import (
	"reflect"
	"testing"
)

func TestParseArgsDefaults(t *testing.T) {
	got, err := ParseArgs(nil)
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	want := Options{}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}

func TestParseArgsReadsFlags(t *testing.T) {
	got, err := ParseArgs([]string{
		"--config", "custom.yaml",
		"--profile", "prod",
		"--readonly",
		"--debug",
	})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	want := Options{
		ConfigPath: "custom.yaml",
		Profile:    "prod",
		ReadOnly:   true,
		Debug:      true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}
