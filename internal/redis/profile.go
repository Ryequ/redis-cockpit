package redis

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ConnectionProfile describes how to connect to one Redis instance or topology.
type ConnectionProfile struct {
	ID                 string   `yaml:"id" json:"id"`
	Name               string   `yaml:"name" json:"name"`
	Addr               string   `yaml:"addr" json:"addr"`
	Username           string   `yaml:"username,omitempty" json:"username,omitempty"`
	Password           string   `yaml:"password,omitempty" json:"password,omitempty"`
	DB                 int      `yaml:"db" json:"db"`
	TLS                bool     `yaml:"tls" json:"tls"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	Cluster            bool     `yaml:"cluster" json:"cluster"`
	Sentinel           bool     `yaml:"sentinel" json:"sentinel"`
	SentinelAddrs      []string `yaml:"sentinel_addrs,omitempty" json:"sentinel_addrs,omitempty"`
	MasterName         string   `yaml:"master_name,omitempty" json:"master_name,omitempty"`
	ReadOnly           bool     `yaml:"read_only" json:"read_only"`
	TimeoutSeconds     int      `yaml:"timeout_seconds" json:"timeout_seconds"`
}

// Expanded returns a copy with credential environment variables resolved.
func (p ConnectionProfile) Expanded() ConnectionProfile {
	p.Username = os.ExpandEnv(p.Username)
	p.Password = os.ExpandEnv(p.Password)
	return p
}

func (p ConnectionProfile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("profile name is required")
	}
	if strings.TrimSpace(p.Addr) == "" {
		return errors.New("profile addr is required")
	}
	if p.DB < 0 {
		return fmt.Errorf("db must be non-negative, got %d", p.DB)
	}
	if p.Cluster && p.Sentinel {
		return errors.New("cluster and sentinel modes are mutually exclusive")
	}
	if p.Sentinel && strings.TrimSpace(p.MasterName) == "" {
		return errors.New("sentinel master name is required")
	}
	if p.TimeoutSeconds < 0 {
		return errors.New("timeout_seconds must be non-negative")
	}
	return nil
}
