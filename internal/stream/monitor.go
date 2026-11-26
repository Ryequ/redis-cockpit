package stream

import (
	"context"
	"errors"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	errNoClient           = errors.New("redis client is not connected")
	errMonitorUnsupported = errors.New("MONITOR is not supported by this client topology")
)

type monitorClient interface {
	Monitor(ctx context.Context, ch chan string) *goredis.MonitorCmd
}

// MonitorEvent is one MONITOR line.
type MonitorEvent struct {
	Payload    string
	ReceivedAt time.Time
}

// Monitor wraps a go-redis MONITOR connection with a bounded queue.
type Monitor struct {
	cmd    *goredis.MonitorCmd
	events *Bounded[MonitorEvent]
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

func (m *Monitor) Events() <-chan MonitorEvent {
	return m.events.C()
}

func (m *Monitor) Dropped() int {
	return m.events.Dropped()
}

func (m *Monitor) Done() <-chan struct{} {
	return m.done
}

func (m *Monitor) Close() error {
	if m == nil {
		return nil
	}
	m.once.Do(func() {
		close(m.done)
		if m.cancel != nil {
			m.cancel()
		}
		if m.cmd != nil {
			m.cmd.Stop()
		}
	})
	return nil
}

// StartMonitor opens a MONITOR stream and captures commands into a bounded queue.
func (h *Hub) StartMonitor(ctx context.Context) (*Monitor, error) {
	if h == nil || h.client == nil {
		return nil, errNoClient
	}
	client, ok := h.client.(monitorClient)
	if !ok {
		return nil, errMonitorUnsupported
	}

	monitorCtx, cancel := context.WithCancel(context.Background())
	lines := make(chan string, 64)
	cmd := client.Monitor(monitorCtx, lines)
	result := &Monitor{
		cmd:    cmd,
		events: NewBounded[MonitorEvent](200),
		cancel: cancel,
		done:   make(chan struct{}),
	}
	cmd.Start()

	go func() {
		for {
			select {
			case <-monitorCtx.Done():
				return
			case line, ok := <-lines:
				if !ok {
					return
				}
				result.events.Send(monitorCtx, MonitorEvent{Payload: line, ReceivedAt: time.Now()})
			}
		}
	}()
	return result, nil
}
