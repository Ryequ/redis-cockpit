package command

import "testing"

func TestRiskOfClassifiesCommands(t *testing.T) {
	tests := []struct {
		argv []string
		want Risk
	}{
		{[]string{"GET", "key"}, RiskRead},
		{[]string{"SET", "key", "value"}, RiskWrite},
		{[]string{"FLUSHALL"}, RiskCritical},
		{[]string{"CONFIG", "SET", "maxmemory", "1gb"}, RiskCritical},
		{[]string{"UNKNOWN"}, RiskNone},
	}

	for _, tt := range tests {
		got := RiskOf(tt.argv)
		if got != tt.want {
			t.Fatalf("RiskOf(%v) = %v, want %v", tt.argv, got, tt.want)
		}
	}
}

func TestRiskOfCoversAdditionalWriteAndCriticalCommands(t *testing.T) {
	tests := []struct {
		argv []string
		want Risk
	}{
		{[]string{"RESTORE", "key", "0", "payload"}, RiskCritical},
		{[]string{"SETNX", "key", "value"}, RiskWrite},
		{[]string{"SCRIPT", "FLUSH"}, RiskCritical},
		{[]string{"CONFIG", "REWRITE"}, RiskCritical},
		{[]string{"MONITOR"}, RiskRead},
	}

	for _, tt := range tests {
		if got := RiskOf(tt.argv); got != tt.want {
			t.Fatalf("RiskOf(%v) = %v, want %v", tt.argv, got, tt.want)
		}
	}
}

func TestRequiresConfirmationForCriticalCommands(t *testing.T) {
	if !RequiresConfirmation([]string{"FLUSHDB"}) {
		t.Fatal("FLUSHDB should require confirmation")
	}
	if !RequiresConfirmation([]string{"EVAL", "return 1"}) {
		t.Fatal("EVAL should require confirmation")
	}
	if RequiresConfirmation([]string{"GET", "key"}) {
		t.Fatal("GET should not require confirmation")
	}
}
