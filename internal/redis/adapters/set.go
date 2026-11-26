package adapters

import "context"

func (m *Manager) SScan(ctx context.Context, key string, cursor uint64, count int64) ([]string, uint64, error) {
	if count <= 0 {
		count = 100
	}
	return m.client().SScan(ctx, key, cursor, "", count).Result()
}
