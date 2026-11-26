package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(_ context.Context, profile ConnectionProfile) (goredis.UniversalClient, error) {
	if err := profile.Validate(); err != nil {
		return nil, err
	}

	timeout := time.Duration(profile.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

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
		sentinelAddrs := sentinelAddresses(profile)
		return goredis.NewFailoverClient(&goredis.FailoverOptions{
			MasterName:    profile.MasterName,
			SentinelAddrs: sentinelAddrs,
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

// TestConnection verifies connectivity and returns the Redis version when available.
func TestConnection(ctx context.Context, profile ConnectionProfile) (string, error) {
	client, err := NewClient(ctx, profile)
	if err != nil {
		return "", err
	}
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}
	info, err := client.Info(ctx, "server").Result()
	if err != nil {
		return "unknown", nil
	}
	return parseRedisVersion(info), nil
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
