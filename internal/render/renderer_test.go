package render

import (
	"strings"
	"testing"
)

func TestRenderLine_Basic(t *testing.T) {
	elems := []RenderedElement{
		{Primary: "branch-main"},
		{Primary: "v1.2.3"},
	}
	got := RenderLine(elems, " | ", 0)
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(got, "branch-main") {
		t.Errorf("output %q missing 'branch-main'", got)
	}
	if !strings.Contains(got, "v1.2.3") {
		t.Errorf("output %q missing 'v1.2.3'", got)
	}
}

func TestRenderLine_WithSeparator(t *testing.T) {
	elems := []RenderedElement{
		{Primary: "left"},
		{Primary: "right"},
	}
	got := RenderLine(elems, " | ", 0)
	if !strings.Contains(got, " | ") {
		t.Errorf("output %q missing separator ' | '", got)
	}
}

func TestRenderLine_CustomSeparator(t *testing.T) {
	elems := []RenderedElement{
		{Primary: "aaa", SeparatorAfter: " -> "},
		{Primary: "bbb"},
		{Primary: "ccc"},
	}
	got := RenderLine(elems, " | ", 0)
	if !strings.Contains(got, " -> ") {
		t.Errorf("output %q missing custom separator ' -> '", got)
	}
	// The separator between bbb and ccc should be the default.
	// Strip the first part up to bbb to check.
	idx := strings.Index(got, "bbb")
	if idx < 0 {
		t.Fatalf("output %q missing 'bbb'", got)
	}
	rest := got[idx+len("bbb"):]
	if !strings.HasPrefix(rest, " | ") {
		t.Errorf("after 'bbb', expected default separator ' | ' but got %q", rest)
	}
}

func TestRenderLine_WidthBudget_Truncation(t *testing.T) {
	elems := []RenderedElement{
		{Primary: "short", Priority: 5},
		{Primary: "this-is-a-very-long-value-that-should-be-truncated", IsVariable: true, MinWidth: 8, Priority: 5},
	}
	// "short | this-is-a-very-long-value-that-should-be-truncated" = 5 + 3 + 50 = 58 display width
	got := RenderLine(elems, " | ", 30)
	stripped := StripANSI(got)
	w := DisplayWidth(stripped)
	if w > 30 {
		t.Errorf("display width %d exceeds budget 30, output: %q", w, stripped)
	}
	if !strings.Contains(stripped, "short") {
		t.Errorf("non-variable element 'short' should still be present, got: %q", stripped)
	}
	if !strings.Contains(stripped, "…") {
		t.Errorf("expected truncation ellipsis in output, got: %q", stripped)
	}
}

func TestRenderLine_WidthBudget_HideByPriority(t *testing.T) {
	elems := []RenderedElement{
		{Primary: "important", Priority: 10},
		{Primary: "optional-detail-that-is-long", Priority: 1},
	}
	// Total: "important | optional-detail-that-is-long" = 9 + 3 + 28 = 40
	// Budget only 12 — must hide the low-priority element.
	got := RenderLine(elems, " | ", 12)
	stripped := StripANSI(got)
	if !strings.Contains(stripped, "important") {
		t.Errorf("high-priority element should remain, got: %q", stripped)
	}
	if strings.Contains(stripped, "optional") {
		t.Errorf("low-priority element should be hidden, got: %q", stripped)
	}
}

func TestRenderLine_NilElements(t *testing.T) {
	got := RenderLine(nil, " | ", 80)
	if got != "" {
		t.Errorf("expected empty string for nil elements, got: %q", got)
	}

	got = RenderLine([]RenderedElement{}, " | ", 80)
	if got != "" {
		t.Errorf("expected empty string for empty slice, got: %q", got)
	}
}

func TestRenderLine_ZeroMaxWidth(t *testing.T) {
	elems := []RenderedElement{
		{Primary: "alpha"},
		{Primary: "beta"},
	}
	got := RenderLine(elems, " | ", 0)
	stripped := StripANSI(got)
	if !strings.Contains(stripped, "alpha") || !strings.Contains(stripped, "beta") {
		t.Errorf("zero maxWidth should mean unlimited, got: %q", stripped)
	}
}

func TestRenderLine_IconAndDetail(t *testing.T) {
	elems := []RenderedElement{
		{Icon: "🔥", Title: "Build", Primary: ":OK", Detail: "(3s)"},
	}
	got := RenderLine(elems, " | ", 0)
	stripped := StripANSI(got)
	// Expected: "🔥 Build:OK (3s)"
	if !strings.Contains(stripped, "🔥 Build:OK (3s)") {
		t.Errorf("unexpected output: %q", stripped)
	}
}

func TestRenderLine_ColorApplied(t *testing.T) {
	elems := []RenderedElement{
		{Primary: "colored", Color: "red"},
	}
	got := RenderLine(elems, "", 0)
	if !strings.Contains(got, "\x1b[31m") {
		t.Errorf("expected ANSI red code in output, got: %q", got)
	}
	if !strings.Contains(got, "\x1b[0m") {
		t.Errorf("expected ANSI reset in output, got: %q", got)
	}
}

// TestRenderAllLines_BlockTheme_UniformWidth verifies that all lines in a
// block-themed output have identical display width due to trailing fill.
func TestRenderAllLines_BlockTheme_UniformWidth(t *testing.T) {
	lines := [][]RenderedElement{
		{
			{Primary: "ABCDE", Color: "#fff", BgColor: "#a00", Priority: 10},
			{Primary: "XY", Color: "#fff", BgColor: "#0a0", Priority: 8},
		},
		{
			{Primary: "AB", Color: "#fff", BgColor: "#a00", Priority: 10},
			{Primary: "XYZW", Color: "#fff", BgColor: "#0a0", Priority: 8},
		},
	}
	result, _ := RenderAllLines(lines, " ", 0)
	if len(result) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result))
	}
	w0 := DisplayWidth(StripANSI(result[0]))
	w1 := DisplayWidth(StripANSI(result[1]))
	if w0 != w1 {
		t.Errorf("lines should have equal display width: %d vs %d", w0, w1)
	}
}

// TestRenderAllLines_PlainTheme_Unchanged verifies non-block themes get
// separator alignment but no trailing fill (widths may still differ).
func TestRenderAllLines_PlainTheme_Unchanged(t *testing.T) {
	lines := [][]RenderedElement{
		{
			{Primary: "ABC", Color: "red", Priority: 10},
			{Primary: "XY", Color: "blue", Priority: 8},
		},
		{
			{Primary: "A", Color: "red", Priority: 10},
		},
	}
	result, _ := RenderAllLines(lines, " | ", 0)
	if len(result) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result))
	}
	// Plain theme: each line should just be elements joined by separator.
	for _, line := range result {
		if line == "" {
			t.Error("expected non-empty line for plain theme")
		}
	}
	// No trailing fill: widths should differ.
	w0 := DisplayWidth(StripANSI(result[0]))
	w1 := DisplayWidth(StripANSI(result[1]))
	if w0 == w1 {
		t.Errorf("plain theme lines should have different widths (no trailing fill), both %d", w0)
	}
}

// TestRenderAllLines_SingleLine verifies single-line input produces no
// alignment or trailing fill.
func TestRenderAllLines_SingleLine(t *testing.T) {
	lines := [][]RenderedElement{
		{
			{Primary: "test", Color: "#fff", BgColor: "#a00", Priority: 10},
		},
	}
	result, _ := RenderAllLines(lines, " ", 0)
	if len(result) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result))
	}
	stripped := StripANSI(result[0])
	if strings.Contains(stripped, "  ") {
		t.Errorf("single line should have no padding spaces, got: %q", stripped)
	}
}

// TestRenderAllLines_MaxWidthRespected verifies all lines respect maxWidth
// even with alignment and trailing fill.
func TestRenderAllLines_MaxWidthRespected(t *testing.T) {
	lines := [][]RenderedElement{
		{
			{Primary: "ABCDEFGH", Color: "#fff", BgColor: "#a00", Priority: 10},
			{Primary: "XY", Color: "#fff", BgColor: "#0a0", Priority: 8},
		},
		{
			{Primary: "A", Color: "#fff", BgColor: "#a00", Priority: 10},
			{Primary: "B", Color: "#fff", BgColor: "#0a0", Priority: 8},
		},
	}
	maxW := 20
	result, _ := RenderAllLines(lines, " ", maxW)
	for i, line := range result {
		w := DisplayWidth(StripANSI(line))
		if w > maxW {
			t.Errorf("line %d width %d exceeds maxWidth %d", i, w, maxW)
		}
	}
}
