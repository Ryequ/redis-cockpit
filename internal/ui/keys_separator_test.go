package ui

import (
	"strings"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestKeysBodyHasSeparatorBetweenListAndValue(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)

	lines := strings.Split(stripANSI(m.renderKeysBody()), "\n")
	keysAt, valueAt := -1, -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "Keys:" {
			keysAt = i
		}
		if strings.HasPrefix(line, "Value:") {
			valueAt = i
		}
	}
	if keysAt < 0 || valueAt < 0 || valueAt <= keysAt {
		t.Fatalf("Keys/Value sections not found in order: keys=%d value=%d", keysAt, valueAt)
	}
	found := false
	for _, line := range lines[keysAt+1 : valueAt] {
		if strings.Contains(line, strings.Repeat("─", 8)) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("separator missing between Keys and Value:\n%s", strings.Join(lines, "\n"))
	}
}
