package ops

import (
	"context"
	"errors"

	"github.com/Ryequ/redis-cockpit/internal/result"
	goredis "github.com/redis/go-redis/v9"
)

var errEmptyCommand = errors.New("empty transaction command")

func (m *Manager) Snapshot(ctx context.Context) (Snapshot, error) {
	client := m.executor.Client()

	infoRaw, err := client.Info(ctx).Result()
	if err != nil {
		return Snapshot{}, err
	}
	info := ParseInfo(infoRaw)
	snapshot := Snapshot{Info: info}

	snapshot.Version = info["Server"]["redis_version"]
	if snapshot.Version == "" {
		snapshot.Version = info["server"]["redis_version"]
	}
	snapshot.Replication = SummarizeReplication(info["Replication"])
	if snapshot.Replication.Role == "" {
		snapshot.Replication = SummarizeReplication(info["replication"])
	}

	if configRaw, err := client.Do(ctx, "CONFIG", "GET", "*").Result(); err == nil {
		snapshot.Config = ParseConfig(configRaw)
	}
	if clientsRaw, err := client.ClientList(ctx).Result(); err == nil {
		snapshot.Clients = ParseClientList(clientsRaw)
	}
	if slowRaw, err := client.SlowLogGet(ctx, 10).Result(); err == nil {
		for _, entry := range slowRaw {
			snapshot.Slowlog = append(snapshot.Slowlog, SlowlogEntry{
				ID:        entry.ID,
				Timestamp: entry.Time.Unix(),
				Duration:  entry.Duration.Milliseconds(),
				Command:   entry.Args,
				Client:    entry.ClientAddr,
				Name:      entry.ClientName,
			})
		}
	}
	if modulesRaw, err := client.Do(ctx, "MODULE", "LIST").Result(); err == nil {
		snapshot.Modules = ParseModules(modulesRaw)
	}
	if aclRaw, err := client.Do(ctx, "ACL", "LIST").Result(); err == nil {
		snapshot.ACL = toStrings(aclRaw)
	}
	return snapshot, nil
}

func (m *Manager) Eval(ctx context.Context, script string, keys []string, args ...any) (result.Value, error) {
	raw, err := m.executor.Client().Eval(ctx, script, keys, args...).Result()
	if err != nil {
		return result.Value{}, err
	}
	return result.Normalize(raw), nil
}

func (m *Manager) Transaction(ctx context.Context, commands [][]any) ([]result.Value, error) {
	client := m.executor.Client()
	pipe := client.TxPipeline()
	cmds := make([]*goredis.Cmd, len(commands))
	for i, argv := range commands {
		if len(argv) == 0 {
			return nil, errEmptyCommand
		}
		cmds[i] = pipe.Do(ctx, argv...)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	values := make([]result.Value, len(cmds))
	for i, cmd := range cmds {
		raw, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		values[i] = result.Normalize(raw)
	}
	return values, nil
}
