package redis

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/codebuddy/redis-tui/internal/command"
)

// RawDoer is implemented by Executor and allows discovery to run COMMAND.
type RawDoer interface {
	Do(ctx context.Context, argv ...any) (any, error)
}

// Discoverer discovers server commands and caches them for a short TTL.
type Discoverer struct {
	doer      RawDoer
	ttl       time.Duration
	cached    []command.Spec
	fetchedAt time.Time
}

func NewDiscoverer(doer RawDoer) *Discoverer {
	return &Discoverer{doer: doer, ttl: 5 * time.Minute}
}

func (d *Discoverer) Discover(ctx context.Context) ([]command.Spec, error) {
	if d.cached != nil && time.Since(d.fetchedAt) < d.ttl {
		return d.cached, nil
	}

	fallback := command.NewRegistry().All()
	if d == nil || d.doer == nil {
		return fallback, nil
	}

	reply, err := d.doer.Do(ctx, "COMMAND")
	if err != nil {
		return fallback, nil
	}

	specs := parseCommandReply(reply)
	if len(specs) == 0 {
		specs = fallback
	}
	d.cached = specs
	d.fetchedAt = time.Now()
	return specs, nil
}

func parseCommandReply(reply any) []command.Spec {
	rv := indirect(reflect.ValueOf(reply))
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil
	}

	specs := make([]command.Spec, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		entry := indirect(rv.Index(i))
		if !entry.IsValid() || (entry.Kind() != reflect.Slice && entry.Kind() != reflect.Array) {
			continue
		}
		if entry.Len() < 3 {
			continue
		}

		name := strings.ToUpper(fmt.Sprint(entry.Index(0).Interface()))
		if name == "" {
			continue
		}
		arity := toInt(entry.Index(1).Interface())
		flags := toStrings(entry.Index(2).Interface())

		specs = append(specs, command.Spec{
			Name:  name,
			Arity: arity,
			Flags: flags,
		})
	}
	return specs
}

func indirect(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func toInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case string:
		var result int
		_, _ = fmt.Sscanf(v, "%d", &result)
		return result
	default:
		return 0
	}
}

func toStrings(value any) []string {
	rv := indirect(reflect.ValueOf(value))
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil
	}
	result := make([]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result = append(result, strings.ToLower(fmt.Sprint(rv.Index(i).Interface())))
	}
	return result
}
