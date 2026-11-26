package command

import (
	"sort"
	"strings"
)

// Registry combines built-in command metadata with dynamically discovered specs.
type Registry struct {
	specs map[string]Spec
}

func NewRegistry() Registry {
	return Registry{specs: map[string]Spec{}}
}

func (r Registry) Has(name string) bool {
	_, ok := r.Find(name)
	return ok
}

func (r Registry) Find(name string) (Spec, bool) {
	key := normalizeName(name)
	if spec, ok := r.specs[key]; ok {
		return spec, true
	}
	if spec, ok := builtinSpecs[key]; ok {
		return spec, true
	}
	return Spec{}, false
}

func (r *Registry) Register(spec Spec) {
	if r.specs == nil {
		r.specs = map[string]Spec{}
	}
	spec.Name = normalizeName(spec.Name)
	r.specs[spec.Name] = spec
}

func (r Registry) Search(query string) []Spec {
	needle := strings.ToLower(strings.TrimSpace(query))
	seen := map[string]bool{}
	results := make([]Spec, 0)

	add := func(spec Spec) {
		if seen[spec.Name] {
			return
		}
		if needle != "" && !strings.Contains(strings.ToLower(spec.Name), needle) && !strings.Contains(strings.ToLower(spec.Summary), needle) {
			return
		}
		seen[spec.Name] = true
		results = append(results, spec)
	}

	for _, spec := range builtinSpecs {
		add(spec)
	}
	for _, spec := range r.specs {
		add(spec)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	return results
}

func (r Registry) All() []Spec {
	return r.Search("")
}

func normalizeName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

var builtinSpecs = buildBuiltins([]Spec{
	{Name: "PING", Arity: -1, Flags: []string{"fast"}, Summary: "Test connection", Category: "server"},
	{Name: "GET", Arity: 2, Flags: []string{"readonly", "fast"}, Summary: "Get string value", Category: "string"},
	{Name: "SET", Arity: -3, Flags: []string{"write"}, Summary: "Set string value", Category: "string"},
	{Name: "MGET", Arity: -2, Flags: []string{"readonly"}, Summary: "Get multiple strings", Category: "string"},
	{Name: "DEL", Arity: -2, Flags: []string{"write"}, Summary: "Delete keys", Category: "key"},
	{Name: "UNLINK", Arity: -2, Flags: []string{"write"}, Summary: "Asynchronously delete keys", Category: "key"},
	{Name: "EXISTS", Arity: -2, Flags: []string{"readonly", "fast"}, Summary: "Check key existence", Category: "key"},
	{Name: "TYPE", Arity: 2, Flags: []string{"readonly", "fast"}, Summary: "Get key type", Category: "key"},
	{Name: "TTL", Arity: 2, Flags: []string{"readonly", "fast"}, Summary: "Get key TTL", Category: "key"},
	{Name: "PTTL", Arity: 2, Flags: []string{"readonly", "fast"}, Summary: "Get key TTL in milliseconds", Category: "key"},
	{Name: "EXPIRE", Arity: -3, Flags: []string{"write"}, Summary: "Set key expiration", Category: "key"},
	{Name: "PERSIST", Arity: 2, Flags: []string{"write"}, Summary: "Remove key expiration", Category: "key"},
	{Name: "SCAN", Arity: -2, Flags: []string{"readonly"}, Summary: "Incrementally scan keys", Category: "key"},
	{Name: "HSET", Arity: -4, Flags: []string{"write"}, Summary: "Set hash fields", Category: "hash"},
	{Name: "HGET", Arity: 3, Flags: []string{"readonly", "fast"}, Summary: "Get hash field", Category: "hash"},
	{Name: "HGETALL", Arity: 2, Flags: []string{"readonly"}, Summary: "Get all hash fields", Category: "hash"},
	{Name: "HSCAN", Arity: -3, Flags: []string{"readonly"}, Summary: "Incrementally scan hash fields", Category: "hash"},
	{Name: "LPUSH", Arity: -3, Flags: []string{"write"}, Summary: "Push to list head", Category: "list"},
	{Name: "RPUSH", Arity: -3, Flags: []string{"write"}, Summary: "Push to list tail", Category: "list"},
	{Name: "LRANGE", Arity: 4, Flags: []string{"readonly"}, Summary: "Get list range", Category: "list"},
	{Name: "SADD", Arity: -3, Flags: []string{"write"}, Summary: "Add set members", Category: "set"},
	{Name: "SMEMBERS", Arity: 2, Flags: []string{"readonly"}, Summary: "Get set members", Category: "set"},
	{Name: "SSCAN", Arity: -3, Flags: []string{"readonly"}, Summary: "Incrementally scan set members", Category: "set"},
	{Name: "ZADD", Arity: -4, Flags: []string{"write"}, Summary: "Add sorted set members", Category: "zset"},
	{Name: "ZRANGE", Arity: -4, Flags: []string{"readonly"}, Summary: "Get sorted set range", Category: "zset"},
	{Name: "ZSCAN", Arity: -3, Flags: []string{"readonly"}, Summary: "Incrementally scan sorted set", Category: "zset"},
	{Name: "XADD", Arity: -5, Flags: []string{"write"}, Summary: "Add stream entry", Category: "stream"},
	{Name: "XRANGE", Arity: -4, Flags: []string{"readonly"}, Summary: "Read stream range", Category: "stream"},
	{Name: "XREAD", Arity: -4, Flags: []string{"readonly", "blocking"}, Summary: "Read from streams", Category: "stream"},
	{Name: "INFO", Arity: -1, Flags: []string{"readonly", "admin"}, Summary: "Get server information", Category: "ops"},
	{Name: "CONFIG", Arity: -2, Flags: []string{"admin"}, Summary: "Read or write configuration", Category: "ops"},
	{Name: "CLIENT", Arity: -2, Flags: []string{"admin"}, Summary: "Manage clients", Category: "ops"},
	{Name: "SLOWLOG", Arity: -2, Flags: []string{"admin"}, Summary: "Inspect slow log", Category: "ops"},
	{Name: "MEMORY", Arity: -2, Flags: []string{"readonly", "admin"}, Summary: "Inspect memory usage", Category: "ops"},
	{Name: "PUBLISH", Arity: 3, Flags: []string{"pubsub"}, Summary: "Publish message", Category: "stream"},
	{Name: "SUBSCRIBE", Arity: -2, Flags: []string{"pubsub", "blocking"}, Summary: "Subscribe to channels", Category: "stream"},
	{Name: "EVAL", Arity: -3, Flags: []string{"scripting"}, Summary: "Execute Lua script", Category: "scripting"},
	{Name: "MULTI", Arity: 1, Flags: []string{"transaction"}, Summary: "Start transaction", Category: "scripting"},
	{Name: "EXEC", Arity: 1, Flags: []string{"transaction"}, Summary: "Execute transaction", Category: "scripting"},
	{Name: "ACL", Arity: -2, Flags: []string{"admin"}, Summary: "Manage ACL users", Category: "security"},
	{Name: "COMMAND", Arity: -1, Flags: []string{"readonly", "admin"}, Summary: "Inspect command metadata", Category: "server"},
	{Name: "MODULE", Arity: -2, Flags: []string{"admin"}, Summary: "Inspect loaded modules", Category: "server"},
	{Name: "FLUSHDB", Arity: -1, Flags: []string{"write", "dangerous"}, Summary: "Flush current database", Category: "danger"},
	{Name: "FLUSHALL", Arity: -1, Flags: []string{"write", "dangerous"}, Summary: "Flush all databases", Category: "danger"},
	{Name: "KEYS", Arity: 2, Flags: []string{"readonly", "dangerous"}, Summary: "List matching keys", Category: "danger"},
})

func buildBuiltins(specs []Spec) map[string]Spec {
	result := make(map[string]Spec, len(specs))
	for _, spec := range specs {
		spec.Name = normalizeName(spec.Name)
		result[spec.Name] = spec
	}
	return result
}
