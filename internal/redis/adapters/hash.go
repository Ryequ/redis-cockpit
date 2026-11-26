package adapters

import "context"

func (m *Manager) HScan(ctx context.Context, key string, cursor uint64, count int64) ([]FieldValue, uint64, error) {
	if count <= 0 {
		count = 100
	}
	values, next, err := m.client().HScan(ctx, key, cursor, "", count).Result()
	if err != nil {
		return nil, 0, err
	}
	fields := make([]FieldValue, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		fields = append(fields, FieldValue{Field: values[i], Value: values[i+1]})
	}
	return fields, next, nil
}
