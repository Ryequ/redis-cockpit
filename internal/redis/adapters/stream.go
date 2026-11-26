package adapters

import (
	"context"
	"fmt"
)

func (m *Manager) XRange(ctx context.Context, key, start, stop string, count int64) ([]StreamEntry, error) {
	messages, err := m.client().XRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	if count > 0 && int64(len(messages)) > count {
		messages = messages[:count]
	}

	entries := make([]StreamEntry, 0, len(messages))
	for _, message := range messages {
		values := make(map[string]string, len(message.Values))
		for field, value := range message.Values {
			values[field] = fmt.Sprint(value)
		}
		entries = append(entries, StreamEntry{ID: message.ID, Values: values})
	}
	return entries, nil
}
