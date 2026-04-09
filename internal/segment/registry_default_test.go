package segment

import "testing"

func TestDefaultRegistry_Count(t *testing.T) {
	r := DefaultRegistry()
	all := r.All()
	if len(all) != 16 {
		t.Fatalf("expected 16 groups, got %d", len(all))
	}
}

func TestDefaultRegistry_NoDuplicateNames(t *testing.T) {
	r := DefaultRegistry()
	seen := map[string]bool{}
	for _, s := range r.All() {
		if seen[s.Name()] {
			t.Fatalf("duplicate segment name: %s", s.Name())
		}
		seen[s.Name()] = true
	}
}

func TestDefaultRegistry_ExpectedGroups(t *testing.T) {
	r := DefaultRegistry()
	expected := []string{
		"model", "git", "context", "tokens", "cost", "usage_5hour", "usage_weekly", "version",
		"session", "speed", "diff", "activity", "live", "cwd", "env", "clock",
	}
	for _, name := range expected {
		if r.Get(name) == nil {
			t.Errorf("expected group %q not found in registry", name)
		}
	}
}
