package security

import (
	"fmt"

	"github.com/Ryequ/redis-cockpit/internal/command"
)

// Policy enforces read-only mode and confirmation requirements.
type Policy struct {
	readOnly bool
	allowed  map[string]bool
}

func NewPolicy(readOnly bool) Policy {
	return Policy{readOnly: readOnly}
}

// WithAllowlist restricts execution to the given command names.
func (p Policy) WithAllowlist(names ...string) Policy {
	p.allowed = make(map[string]bool, len(names))
	for _, name := range names {
		p.allowed[name] = true
	}
	return p
}

func (p Policy) Check(argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("empty command")
	}
	if p.allowed != nil && !p.allowed[normalize(argv[0])] {
		return fmt.Errorf("command %s is not in the allowlist", argv[0])
	}
	if p.readOnly && command.RiskOf(argv) != command.RiskRead {
		return fmt.Errorf("command %s is blocked in read-only mode", argv[0])
	}
	return nil
}

func (p Policy) RequiresConfirmation(argv []string) bool {
	return command.RequiresConfirmation(argv)
}

func normalize(name string) string {
	result := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
