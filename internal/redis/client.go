package redis

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, profile ConnectionProfile) (goredis.UniversalClient, error) {
	if profile.SSHHost != "" {
		return nil, fmt.Errorf("ssh profile requires Open")
	}
	return newDirectClient(ctx, profile)
}

// Open creates a Redis client and an optional closer for an SSH tunnel.
func Open(ctx context.Context, profile ConnectionProfile) (goredis.UniversalClient, func() error, error) {
	if err := profile.Validate(); err != nil {
		return nil, nil, err
	}
	if profile.SSHHost == "" {
		client, err := newDirectClient(ctx, profile)
		return client, nil, err
	}

	localPort, err := freeLocalPort()
	if err != nil {
		return nil, nil, err
	}
	cmd := exec.Command("ssh", sshTunnelArgs(profile, localPort)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start ssh tunnel: %w", err)
	}
	stopTunnel := func() error {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		return nil
	}

	localAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(localPort))
	if err := waitForPort(ctx, localAddr, dialTimeout(profile)); err != nil {
		_ = stopTunnel()
		return nil, nil, fmt.Errorf("ssh tunnel not ready: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	tunneled := profile
	tunneled.Addr = localAddr
	client, err := newDirectClient(ctx, tunneled)
	if err != nil {
		_ = stopTunnel()
		return nil, nil, err
	}
	closer := func() error {
		_ = client.Close()
		return stopTunnel()
	}
	return client, closer, nil
}

func newDirectClient(_ context.Context, profile ConnectionProfile) (goredis.UniversalClient, error) {
	if err := profile.Validate(); err != nil {
		return nil, err
	}

	timeout := dialTimeout(profile)
	tlsConfig := func() *tls.Config {
		if !profile.TLS {
			return nil
		}
		return &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: profile.InsecureSkipVerify, // #nosec G402 -- explicit user setting
		}
	}()

	if profile.Cluster {
		return goredis.NewClusterClient(&goredis.ClusterOptions{
			Addrs:        []string{profile.Addr},
			Username:     profile.Username,
			Password:     profile.Password,
			TLSConfig:    tlsConfig,
			DialTimeout:  timeout,
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		}), nil
	}

	if profile.Sentinel {
		return goredis.NewFailoverClient(&goredis.FailoverOptions{
			MasterName:    profile.MasterName,
			SentinelAddrs: sentinelAddresses(profile),
			Username:      profile.Username,
			Password:      profile.Password,
			DB:            profile.DB,
			TLSConfig:     tlsConfig,
			DialTimeout:   timeout,
			ReadTimeout:   timeout,
			WriteTimeout:  timeout,
		}), nil
	}

	return goredis.NewClient(&goredis.Options{
		Addr:         profile.Addr,
		Username:     profile.Username,
		Password:     profile.Password,
		DB:           profile.DB,
		TLSConfig:    tlsConfig,
		DialTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}), nil
}

func dialTimeout(profile ConnectionProfile) time.Duration {
	timeout := time.Duration(profile.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		return 5 * time.Second
	}
	return timeout
}

func freeLocalPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func waitForPort(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return fmt.Errorf("timeout waiting for %s", addr)
}

func sshTunnelArgs(profile ConnectionProfile, localPort int) []string {
	sshPort := profile.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}
	host, port := splitRedisAddress(profile.Addr)
	args := []string{
		"-N",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ExitOnForwardFailure=yes",
		"-L", fmt.Sprintf("127.0.0.1:%d:%s:%s", localPort, host, port),
		"-p", strconv.Itoa(sshPort),
	}
	if strings.TrimSpace(profile.SSHKeyPath) != "" {
		args = append(args, "-i", profile.SSHKeyPath)
	}
	return append(args, profile.SSHUser+"@"+profile.SSHHost)
}

func splitRedisAddress(addr string) (string, string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil || host == "" {
		if err != nil {
			host = addr
		}
		port = "6379"
	}
	return host, port
}

// Probe verifies connectivity and returns Redis version and ping latency.
func Probe(ctx context.Context, profile ConnectionProfile) (string, time.Duration, error) {
	client, closer, err := Open(ctx, profile)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		if closer != nil {
			_ = closer()
		} else {
			_ = client.Close()
		}
	}()

	start := time.Now()
	if err := client.Ping(ctx).Err(); err != nil {
		return "", time.Since(start), fmt.Errorf("ping failed: %w", err)
	}
	latency := time.Since(start)
	if latency <= 0 {
		latency = time.Nanosecond
	}

	version := "unknown"
	if info, err := client.Info(ctx, "server").Result(); err == nil {
		version = parseRedisVersion(info)
	}
	return version, latency, nil
}

// TestConnection verifies connectivity and returns the Redis version when available.
func TestConnection(ctx context.Context, profile ConnectionProfile) (string, error) {
	version, _, err := Probe(ctx, profile)
	return version, err
}

func sentinelAddresses(profile ConnectionProfile) []string {
	if len(profile.SentinelAddrs) > 0 {
		return append([]string{}, profile.SentinelAddrs...)
	}
	if profile.Addr != "" {
		return []string{profile.Addr}
	}
	return nil
}

func parseRedisVersion(info string) string {
	const marker = "redis_version:"
	for i := 0; i+len(marker) <= len(info); i++ {
		if info[i:i+len(marker)] == marker {
			rest := info[i+len(marker):]
			end := 0
			for end < len(rest) && rest[end] != '\r' && rest[end] != '\n' {
				end++
			}
			return rest[:end]
		}
	}
	return "unknown"
}
