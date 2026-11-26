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
	SSHHost            string   `yaml:"ssh_host,omitempty" json:"ssh_host,omitempty"`
	SSHPort            int      `yaml:"ssh_port,omitempty" json:"ssh_port,omitempty"`
	SSHUser            string   `yaml:"ssh_user,omitempty" json:"ssh_user,omitempty"`
	SSHKeyPath         string   `yaml:"ssh_key_path,omitempty" json:"ssh_key_path,omitempty"`
	CloudProvider      string   `yaml:"cloud_provider,omitempty" json:"cloud_provider,omitempty"`
	AuthMechanism      string   `yaml:"auth_mechanism,omitempty" json:"auth_mechanism,omitempty"`
}

func ApplyCloudPreset(profile ConnectionProfile, provider string) ConnectionProfile {
	profile.CloudProvider = strings.ToLower(strings.TrimSpace(provider))
	profile.TLS = true
	switch profile.CloudProvider {
	case "aws":
		profile.AuthMechanism = "iam"
	case "gcp", "azure", "aliyun", "tencent":
		profile.AuthMechanism = "default"
	default:
		profile.AuthMechanism = "default"
	}
	return profile
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
	if p.SSHHost != "" && strings.TrimSpace(p.SSHUser) == "" {
		return errors.New("ssh user is required when ssh_host is set")
	}
	if p.SSHPort < 0 {
		return fmt.Errorf("ssh_port must be non-negative, got %d", p.SSHPort)
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
