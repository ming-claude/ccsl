package segment

import "testing"

// stubSegment is a minimal Segment implementation for testing.
type stubSegment struct {
	name     string
	priority int
}

func (s *stubSegment) Name() string                                    { return s.name }
func (s *stubSegment) Render(_ *RenderContext) (*SegmentOutput, error) { return nil, nil }
func (s *stubSegment) DefaultTitle() string                            { return s.name }
func (s *stubSegment) DefaultIcon(_ StyleMode) string                  { return "" }
func (s *stubSegment) DefaultPriority() int                            { return s.priority }

func TestRegistryLookup(t *testing.T) {
	reg := NewRegistry()
	stub := &stubSegment{name: "model", priority: 1}
	reg.Register(stub)

	got := reg.Get("model")
	if got == nil {
		t.Fatal("expected segment 'model', got nil")
	}
	if got.Name() != "model" {
		t.Fatalf("expected name 'model', got %q", got.Name())
	}

	if reg.Get("unknown") != nil {
		t.Fatal("expected nil for unknown segment")
	}
}

func TestRegistryAll(t *testing.T) {
	reg := NewRegistry()
	s1 := &stubSegment{name: "alpha", priority: 1}
	s2 := &stubSegment{name: "beta", priority: 2}
	reg.Register(s1)
	reg.Register(s2)

	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(all))
	}
	if all[0].Name() != "alpha" {
		t.Fatalf("expected first segment 'alpha', got %q", all[0].Name())
	}
	if all[1].Name() != "beta" {
		t.Fatalf("expected second segment 'beta', got %q", all[1].Name())
	}
}
