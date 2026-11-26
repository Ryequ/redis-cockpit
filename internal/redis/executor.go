package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/codebuddy/redis-tui/internal/result"
	goredis "github.com/redis/go-redis/v9"
)

// Executor runs arbitrary Redis commands through a UniversalClient.
type Executor struct {
	client goredis.UniversalClient
}

func NewExecutor(client goredis.UniversalClient) *Executor {
	return &Executor{client: client}
}

func (e *Executor) Execute(ctx context.Context, argv []any) (result.Value, error) {
	if len(argv) == 0 {
		return result.Value{}, errors.New("empty command")
	}
	raw, err := e.Do(ctx, argv...)
	if err != nil {
		return result.Value{}, err
	}
	return result.Normalize(raw), nil
}

// Do exposes the raw reply for discovery and specialized adapters.
func (e *Executor) Do(ctx context.Context, argv ...any) (any, error) {
	if e == nil || e.client == nil {
		return nil, errors.New("redis client is not connected")
	}
	if len(argv) == 0 {
		return nil, errors.New("empty command")
	}
	return e.client.Do(ctx, argv...).Result()
}

func (e *Executor) Client() goredis.UniversalClient {
	return e.client
}

func (e *Executor) Close() error {
	if e == nil || e.client == nil {
		return nil
	}
	if err := e.client.Close(); err != nil {
		return fmt.Errorf("close redis client: %w", err)
	}
	return nil
}
