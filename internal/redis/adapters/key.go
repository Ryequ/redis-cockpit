package adapters

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

func (m *Manager) ScanKeys(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	if count <= 0 {
		count = 100
	}
	if match == "" {
		match = "*"
	}
	return m.client().Scan(ctx, cursor, match, count).Result()
}

func (m *Manager) DescribeKeys(ctx context.Context, keys []string) ([]KeyInfo, error) {
	client := m.client()
	if len(keys) == 0 {
		return nil, nil
	}

	pipe := client.Pipeline()
	typeCmds := make([]*goredis.StatusCmd, len(keys))
	ttlCmds := make([]*goredis.DurationCmd, len(keys))
	for i, key := range keys {
		typeCmds[i] = pipe.Type(ctx, key)
		ttlCmds[i] = pipe.TTL(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	infos := make([]KeyInfo, len(keys))
	for i, key := range keys {
		infos[i] = KeyInfo{Name: key, Type: typeCmds[i].Val(), TTL: ttlCmds[i].Val()}
	}
	return infos, nil
}

func (m *Manager) DeleteKeys(ctx context.Context, keys []string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return m.client().Del(ctx, keys...).Result()
}
