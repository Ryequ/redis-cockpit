package adapters

import (
	"context"
	"time"
)

func (m *Manager) GetString(ctx context.Context, key string) (string, error) {
	return m.client().Get(ctx, key).Result()
}

func (m *Manager) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	return m.client().Set(ctx, key, value, ttl).Err()
}
