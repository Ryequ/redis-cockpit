package command

import "testing"

func TestRegistryFindsBuiltinCommands(t *testing.T) {
	r := NewRegistry()

	spec, ok := r.Find("GET")
	if !ok {
		t.Fatal("GET not found")
	}
	if spec.Name != "GET" {
		t.Fatalf("spec name = %q, want GET", spec.Name)
	}

	for _, name := range []string{"SET", "SCAN", "HSET", "XRANGE", "INFO"} {
		if !r.Has(name) {
			t.Fatalf("%s should be registered", name)
		}
	}
}

func TestRegistrySearchIsCaseInsensitive(t *testing.T) {
	r := NewRegistry()
	results := r.Search("set")

	if len(results) == 0 {
		t.Fatal("search returned no results")
	}
	for _, spec := range results {
		if !containsFold(spec.Name, "set") && !containsFold(spec.Summary, "set") {
			t.Fatalf("result %q does not match set", spec.Name)
		}
	}
}

func containsFold(s, sub string) bool {
	return len(s) >= len(sub) && (lowerContains(s, sub))
}

func lowerContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if lowerByte(s[i+j]) != lowerByte(sub[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func lowerByte(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
