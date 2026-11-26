package security

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// AuditEvent is one redacted command execution record.
type AuditEvent struct {
	Time    time.Time `json:"time"`
	Profile string    `json:"profile"`
	Command string    `json:"command"`
}

// AuditLog appends redacted command records to a JSON Lines file.
type AuditLog struct {
	path string
	mu   sync.Mutex
}

func NewAuditLog(path string) *AuditLog {
	return &AuditLog{path: path}
}

func (l *AuditLog) Record(profile string, argv []string) error {
	if l == nil || l.path == "" {
		return nil
	}

	event := AuditEvent{
		Time:    time.Now().UTC(),
		Profile: profile,
		Command: strings.Join(RedactCommand(argv), " "),
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}
