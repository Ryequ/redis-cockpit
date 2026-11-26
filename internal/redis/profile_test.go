package redis

import "testing"

func TestProfileValidateAcceptsSingleInstance(t *testing.T) {
	p := ConnectionProfile{Name: "local", Addr: "127.0.0.1:6379"}
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestProfileValidateRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		profile ConnectionProfile
		want    string
	}{
		{"missing name", ConnectionProfile{Addr: "127.0.0.1:6379"}, "name"},
		{"missing addr", ConnectionProfile{Name: "local"}, "addr"},
		{"invalid db", ConnectionProfile{Name: "local", Addr: "127.0.0.1:6379", DB: -1}, "db"},
		{
			"cluster and sentinel conflict",
			ConnectionProfile{
				Name:       "local",
				Addr:       "127.0.0.1:6379",
				Cluster:    true,
				Sentinel:   true,
				MasterName: "mymaster",
			},
			"cluster",
		},
		{
			"sentinel requires master name",
			ConnectionProfile{
				Name:     "local",
				Addr:     "127.0.0.1:26379",
				Sentinel: true,
			},
			"master name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if err == nil {
				t.Fatal("Validate returned nil, want error")
			}
			if !contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.want)
			}
		})
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
