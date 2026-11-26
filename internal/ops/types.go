package ops

import redisclient "github.com/codebuddy/redis-tui/internal/redis"

// ClientInfo is a parsed CLIENT LIST row.
type ClientInfo struct {
	ID   string
	Addr string
	Name string
	DB   string
	Cmd  string
}

// SlowlogEntry is a parsed SLOWLOG GET entry.
type SlowlogEntry struct {
	ID        int64
	Timestamp int64
	Duration  int64
	Command   []string
	Client    string
	Name      string
}

// ReplicationInfo summarizes replication state.
type ReplicationInfo struct {
	Role            string
	MasterHost      string
	MasterPort      string
	ConnectedSlaves int
}

// Snapshot is the data shown in the Ops view.
type Snapshot struct {
	Info        map[string]map[string]string
	Config      map[string]string
	Clients     []ClientInfo
	Slowlog     []SlowlogEntry
	Replication ReplicationInfo
	Modules     []string
	ACL         []string
	Version     string
}

// Manager reads ops data and runs scripting/transaction commands.
type Manager struct {
	executor *redisclient.Executor
}

func New(executor *redisclient.Executor) *Manager {
	return &Manager{executor: executor}
}
