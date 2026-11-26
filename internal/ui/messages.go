package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/Ryequ/redis-cockpit/internal/ops"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
	"github.com/Ryequ/redis-cockpit/internal/result"
	"github.com/Ryequ/redis-cockpit/internal/security"
	"github.com/Ryequ/redis-cockpit/internal/stream"
)

type connectedMsg struct {
	gen      uint64
	profile  redis.ConnectionProfile
	executor *redis.Executor
	close    func() error
	err      error
}

type profileTestedMsg struct {
	gen        uint64
	profileGen uint64
	name       string
	latency    time.Duration
	version    string
	err        error
}

type commandExecutedMsg struct {
	gen   uint64
	line  string
	value result.Value
	err   error
}

func connectCmd(profile redis.ConnectionProfile, gen uint64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), profileTimeout(profile))
		defer cancel()

		client, closeClient, err := redis.Open(ctx, profile.Expanded())
		if err != nil {
			return connectedMsg{gen: gen, err: err}
		}
		cleanup := func() error {
			if closeClient != nil {
				return closeClient()
			}
			return client.Close()
		}
		if err := client.Ping(ctx).Err(); err != nil {
			_ = cleanup()
			return connectedMsg{gen: gen, err: fmt.Errorf("connect to %s: %w", profile.Addr, err)}
		}
		return connectedMsg{gen: gen, profile: profile, executor: redis.NewExecutor(client), close: cleanup}
	}
}

func testConnectionCmd(profile redis.ConnectionProfile, gen, profileGen uint64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), profileTimeout(profile))
		defer cancel()

		version, latency, err := redis.Probe(ctx, profile.Expanded())
		return profileTestedMsg{
			gen:        gen,
			profileGen: profileGen,
			name:       profile.Name,
			latency:    latency,
			version:    version,
			err:        err,
		}
	}
}

func profileTimeout(profile redis.ConnectionProfile) time.Duration {
	timeout := time.Duration(profile.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		return 5 * time.Second
	}
	return timeout
}

func executeCmd(executor *redis.Executor, argv []string, gen uint64) tea.Cmd {
	line := strings.Join(security.RedactCommand(argv), " ")
	return func() tea.Msg {
		if executor == nil {
			return commandExecutedMsg{gen: gen, line: line, err: errors.New("not connected to Redis")}
		}
		args := make([]any, len(argv))
		for i, arg := range argv {
			args[i] = arg
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		value, err := executor.Execute(ctx, args)
		return commandExecutedMsg{gen: gen, line: line, value: value, err: err}
	}
}

type keysLoadedMsg struct {
	gen    uint64
	infos  []adapters.KeyInfo
	cursor uint64
	err    error
}

type valueLoadedMsg struct {
	gen   uint64
	key   string
	value result.Value
	err   error
}

type keysDeletedMsg struct {
	gen     uint64
	keys    []string
	deleted int64
	err     error
}

func loadKeysCmd(manager *adapters.Manager, match string, count int64, cursor, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			return keysLoadedMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		keys, next, err := manager.ScanKeys(ctx, cursor, match, count)
		if err != nil {
			return keysLoadedMsg{gen: gen, err: err}
		}
		infos, err := manager.DescribeKeys(ctx, keys)
		if err != nil {
			return keysLoadedMsg{gen: gen, err: err}
		}
		return keysLoadedMsg{gen: gen, infos: infos, cursor: next}
	}
}

func loadValueCmd(manager *adapters.Manager, info adapters.KeyInfo, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			return valueLoadedMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		value, err := loadTypedValue(ctx, manager, info)
		return valueLoadedMsg{gen: gen, key: info.Name, value: value, err: err}
	}
}

func deleteKeysCmd(manager *adapters.Manager, keys []string, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			return keysDeletedMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		deleted, err := manager.DeleteKeys(ctx, keys)
		return keysDeletedMsg{gen: gen, keys: keys, deleted: deleted, err: err}
	}
}

type valueCopiedMsg struct {
	gen uint64
	err error
}

type clipboardCommandSpec struct {
	name string
	args []string
}

var (
	newClipboardCommand           = clipboardCommand
	clipboardWriter     io.Writer = os.Stdout
)

func copyValueCmd(value string, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if spec := newClipboardCommand(value); spec != nil {
			cmd := exec.Command(spec.name, spec.args...)
			cmd.Stdin = strings.NewReader(value)
			var stderr strings.Builder
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil {
				return valueCopiedMsg{gen: gen}
			}
		}
		if _, err := io.WriteString(clipboardWriter, ansi.SetSystemClipboard(value)); err != nil {
			return valueCopiedMsg{gen: gen, err: fmt.Errorf("write OSC52 clipboard: %w", err)}
		}
		return valueCopiedMsg{gen: gen}
	}
}

func clipboardCommand(string) *clipboardCommandSpec {
	candidates := []clipboardCommandSpec{
		{name: "wl-copy"},
		{name: "xclip", args: []string{"-selection", "clipboard"}},
		{name: "xsel", args: []string{"--input", "--clipboard"}},
	}
	if runtime.GOOS == "darwin" {
		candidates = append([]clipboardCommandSpec{{name: "pbcopy"}}, candidates...)
	}
	if runtime.GOOS == "windows" {
		candidates = append([]clipboardCommandSpec{{name: "clip"}}, candidates...)
	}
	for _, spec := range candidates {
		if _, err := exec.LookPath(spec.name); err == nil {
			return &spec
		}
	}
	return nil
}

type keyWrittenMsg struct {
	gen uint64
	key string
	err error
}

func createKeyCmd(manager *adapters.Manager, key, value string, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			return keyWrittenMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := manager.SetString(ctx, key, value, 0)
		return keyWrittenMsg{gen: gen, key: key, err: err}
	}
}

func updateKeyValueCmd(manager *adapters.Manager, key, value string, ttl time.Duration, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			return keyWrittenMsg{gen: gen, key: key, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if ttl < 0 {
			ttl = 0
		}
		err := manager.SetString(ctx, key, value, ttl)
		return keyWrittenMsg{gen: gen, key: key, err: err}
	}
}

func loadTypedValue(ctx context.Context, manager *adapters.Manager, info adapters.KeyInfo) (result.Value, error) {
	switch info.Type {
	case "string":
		value, err := manager.GetString(ctx, info.Name)
		return result.Normalize(value), err
	case "hash":
		fields, _, err := manager.HScan(ctx, info.Name, 0, 100)
		if err != nil {
			return result.Value{}, err
		}
		items := make([]any, len(fields))
		for i, field := range fields {
			items[i] = map[string]string{"field": field.Field, "value": field.Value}
		}
		return result.Normalize(items), nil
	case "list":
		items, err := manager.LRange(ctx, info.Name, 0, 199)
		if err != nil {
			return result.Value{}, err
		}
		return result.Normalize(toAnySlice(items)), nil
	case "set":
		members, _, err := manager.SScan(ctx, info.Name, 0, 100)
		if err != nil {
			return result.Value{}, err
		}
		return result.Normalize(toAnySlice(members)), nil
	case "zset":
		members, _, err := manager.ZScan(ctx, info.Name, 0, 100)
		if err != nil {
			return result.Value{}, err
		}
		items := make([]any, len(members))
		for i, member := range members {
			items[i] = map[string]any{"member": member.Member, "score": member.Score}
		}
		return result.Normalize(items), nil
	case "stream":
		entries, err := manager.XRange(ctx, info.Name, "-", "+", 100)
		if err != nil {
			return result.Value{}, err
		}
		items := make([]any, len(entries))
		for i, entry := range entries {
			items[i] = map[string]any{"id": entry.ID, "values": entry.Values}
		}
		return result.Normalize(items), nil
	default:
		return result.Value{}, fmt.Errorf("unsupported key type %q", info.Type)
	}
}

func toAnySlice(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}

type opsLoadedMsg struct {
	gen      uint64
	snapshot ops.Snapshot
	err      error
}

type subscribedMsg struct {
	gen          uint64
	subscription *stream.Subscription
	err          error
}

type publishedMsg struct {
	gen uint64
	err error
}

type monitorStartedMsg struct {
	gen     uint64
	monitor *stream.Monitor
	err     error
}

type streamMessageMsg struct {
	gen     uint64
	message stream.Message
}

type monitorEventMsg struct {
	gen   uint64
	event stream.MonitorEvent
}

type streamClosedMsg struct {
	gen uint64
}

func loadOpsCmd(manager *ops.Manager, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if manager == nil {
			return opsLoadedMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		snapshot, err := manager.Snapshot(ctx)
		return opsLoadedMsg{gen: gen, snapshot: snapshot, err: err}
	}
}

func subscribeCmd(hub *stream.Hub, channel string, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if hub == nil {
			return subscribedMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		subscription, err := hub.Subscribe(ctx, channel)
		return subscribedMsg{gen: gen, subscription: subscription, err: err}
	}
}

func publishCmd(hub *stream.Hub, channel, payload string, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if hub == nil {
			return publishedMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return publishedMsg{gen: gen, err: hub.Publish(ctx, channel, payload)}
	}
}

func startMonitorCmd(hub *stream.Hub, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if hub == nil {
			return monitorStartedMsg{gen: gen, err: errors.New("not connected to Redis")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		monitor, err := hub.StartMonitor(ctx)
		return monitorStartedMsg{gen: gen, monitor: monitor, err: err}
	}
}

func waitStreamMessageCmd(subscription *stream.Subscription, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if subscription == nil {
			return streamClosedMsg{gen: gen}
		}
		select {
		case message := <-subscription.Messages():
			return streamMessageMsg{gen: gen, message: message}
		case <-subscription.Done():
			return streamClosedMsg{gen: gen}
		}
	}
}

func waitMonitorEventCmd(monitor *stream.Monitor, gen uint64) tea.Cmd {
	return func() tea.Msg {
		if monitor == nil {
			return streamClosedMsg{gen: gen}
		}
		select {
		case event := <-monitor.Events():
			return monitorEventMsg{gen: gen, event: event}
		case <-monitor.Done():
			return streamClosedMsg{gen: gen}
		}
	}
}
