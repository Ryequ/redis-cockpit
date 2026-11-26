package adapters

import (
	"context"
	"strconv"
)

func (m *Manager) ZScan(ctx context.Context, key string, cursor uint64, count int64) ([]MemberScore, uint64, error) {
	if count <= 0 {
		count = 100
	}
	values, next, err := m.client().ZScan(ctx, key, cursor, "", count).Result()
	if err != nil {
		return nil, 0, err
	}

	members := make([]MemberScore, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		score, _ := strconv.ParseFloat(values[i+1], 64)
		members = append(members, MemberScore{Member: values[i], Score: score})
	}
	return members, next, nil
}
