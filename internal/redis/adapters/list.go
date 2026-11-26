package adapters

import "context"

func (m *Manager) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return m.client().LRange(ctx, key, start, stop).Result()
}
