package command

import "strings"

// Risk expresses how dangerous a command is.
type Risk int

const (
	RiskNone Risk = iota
	RiskRead
	RiskWrite
	RiskCritical
)

func (r Risk) String() string {
	switch r {
	case RiskRead:
		return "read"
	case RiskWrite:
		return "write"
	case RiskCritical:
		return "critical"
	default:
		return "none"
	}
}

var readCommands = setOf(
	"GET", "MGET", "EXISTS", "TYPE", "TTL", "PTTL", "SCAN", "HGET", "HMGET",
	"HGETALL", "HSCAN", "HLEN", "LRANGE", "LLEN", "SMEMBERS", "SSCAN", "SCARD",
	"SISMEMBER", "ZRANGE", "ZSCAN", "ZCARD", "ZSCORE", "XRANGE", "XREVRANGE",
	"XLEN", "INFO", "CONFIG", "CLIENT", "SLOWLOG", "COMMAND", "MODULE", "PING",
	"ACL", "MEMORY", "LATENCY", "ROLE", "DBSIZE", "MONITOR", "WAIT",
	"XREAD", "XINFO", "XPENDING", "STRLEN", "GETRANGE", "HKEYS", "HVALS",
	"LINDEX", "ZRANK", "ZREVRANK", "ZCOUNT", "ZLEXCOUNT", "PFCOUNT", "BITCOUNT",
	"GETBIT", "TYPE", "DUMP", "SCAN",
)

var writeCommands = setOf(
	"SET", "SETEX", "PSETEX", "SETNX", "GETSET", "GETDEL", "APPEND",
	"INCR", "INCRBY", "INCRBYFLOAT", "DECR", "DECRBY", "HSET", "HSETNX", "HDEL",
	"LPUSH", "RPUSH", "LPOP", "RPOP", "LSET", "LINSERT", "LTRIM", "LPOS",
	"SADD", "SREM", "SMOVE", "ZADD", "ZREM", "ZINCRBY", "ZPOPMIN", "ZPOPMAX",
	"XADD", "XDEL", "XACK", "XCLAIM", "XAUTOCLAIM", "XGROUP",
	"SINTERSTORE", "SUNIONSTORE", "SDIFFSTORE", "ZINTERSTORE", "ZUNIONSTORE",
	"EXPIRE", "PEXPIRE", "EXPIREAT", "PEXPIREAT", "PERSIST", "RENAME", "RENAMENX",
	"MULTI", "EXEC", "DISCARD", "WATCH", "UNWATCH", "PUBLISH", "SELECT",
	"SUBSCRIBE", "UNSUBSCRIBE", "PSUBSCRIBE", "PUNSUBSCRIBE", "COPY", "MOVE",
	"BITSET", "BITFIELD", "SETBIT", "SETRANGE", "PFADD", "PFMERGE",
)

var criticalCommands = setOf(
	"DEL", "UNLINK", "FLUSHDB", "FLUSHALL", "KEYS", "SHUTDOWN", "DEBUG",
	"CONFIG", "ACL", "SENTINEL", "REPLICAOF", "SLAVEOF", "BGREWRITEAOF",
	"BGSAVE", "SAVE", "SWAPDB", "RESTORE", "MIGRATE", "EVAL", "EVALSHA",
	"SCRIPT", "FUNCTION", "CLIENT", "SLOWLOG", "MEMORY", "LATENCY",
)

// RiskOf classifies an argv slice by command name and selected subcommand.
func RiskOf(argv []string) Risk {
	if len(argv) == 0 {
		return RiskNone
	}
	name := strings.ToUpper(strings.TrimSpace(argv[0]))

	switch name {
	case "CONFIG":
		if len(argv) >= 2 && containsFoldAny([]string{"SET", "REWRITE"}, argv[1]) {
			return RiskCritical
		}
		return RiskRead
	case "ACL":
		if len(argv) >= 2 && containsFoldAny([]string{"DELETE", "SETUSER", "DELUSER", "LOAD"}, argv[1]) {
			return RiskCritical
		}
		return RiskRead
	case "CLIENT":
		if len(argv) >= 2 && containsFoldAny([]string{"KILL", "PAUSE", "NO-EVICT", "TRACKING"}, argv[1]) {
			return RiskCritical
		}
		return RiskRead
	case "SLOWLOG":
		if len(argv) >= 2 && strings.EqualFold(argv[1], "RESET") {
			return RiskCritical
		}
		return RiskRead
	case "MEMORY":
		if len(argv) >= 2 && strings.EqualFold(argv[1], "PURGE") {
			return RiskCritical
		}
		return RiskRead
	case "LATENCY":
		if len(argv) >= 2 && strings.EqualFold(argv[1], "RESET") {
			return RiskCritical
		}
		return RiskRead
	case "SCRIPT":
		if len(argv) >= 2 && strings.EqualFold(argv[1], "FLUSH") {
			return RiskCritical
		}
		return RiskCritical
	case "FUNCTION":
		if len(argv) >= 2 && containsFoldAny([]string{"DELETE", "FLUSH", "RESTORE", "LOAD"}, argv[1]) {
			return RiskCritical
		}
		return RiskCritical
	}

	if criticalCommands[name] {
		return RiskCritical
	}
	if writeCommands[name] {
		return RiskWrite
	}
	if readCommands[name] {
		return RiskRead
	}
	return RiskNone
}

// RequiresConfirmation reports whether the command should show a modal.
func RequiresConfirmation(argv []string) bool {
	return RiskOf(argv) == RiskCritical
}

func containsFoldAny(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func setOf(values ...string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
