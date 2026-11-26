package adapters

import (
	"time"

	goredis "github.com/redis/go-redis/v9"

	redisclient "github.com/Ryequ/redis-cockpit/internal/redis"
)

// KeyInfo is the metadata shown in the key browser.
type KeyInfo struct {
	Name string
	Type string
	TTL  time.Duration
}

// FieldValue is one hash field.
type FieldValue struct {
	Field string
	Value string
}

// MemberScore is one sorted-set member.
type MemberScore struct {
	Member string
	Score  float64
}

// StreamEntry is one stream message.
type StreamEntry struct {
	ID     string
	Values map[string]string
}

// Manager groups typed adapters for the core Redis data types.
type Manager struct {
	executor *redisclient.Executor
}

func New(executor *redisclient.Executor) *Manager {
	return &Manager{executor: executor}
}

func (m *Manager) client() goredis.UniversalClient {
	if m == nil || m.executor == nil {
		return nil
	}
	return m.executor.Client()
}
