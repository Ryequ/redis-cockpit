package config

import (
	"fmt"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

// UpsertProfile adds or replaces a profile by name.
func (c *Config) UpsertProfile(profile redis.ConnectionProfile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("validate profile: %w", err)
	}
	for i, existing := range c.Profiles {
		if existing.Name == profile.Name {
			c.Profiles[i] = profile
			return nil
		}
	}
	c.Profiles = append(c.Profiles, profile)
	return nil
}

// DeleteProfile removes a profile by name.
func (c *Config) DeleteProfile(name string) bool {
	for i, profile := range c.Profiles {
		if profile.Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			if c.ActiveProfile == name {
				c.ActiveProfile = ""
			}
			return true
		}
	}
	return false
}
