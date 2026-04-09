package render

import (
	"strings"
	"testing"
)

// testElemWidths computes elemWidths from elements (simulating un-truncated budget).
func testElemWidths(allLines [][]RenderedElement) [][]int {
	widths := make([][]int, len(allLines))
	for li, line := range allLines {
		widths[li] = make([]int, len(line))
		for ei, e := range line {
			w := DisplayWidth(buildElementText(e))
			if e.BgColor != "" {
				w += 2 // block padding
			}
			widths[li][ei] = w
		}
	}
	return widths
}

// TestMergeSameColorBlocks: 2 teal elements + 1 gold element → 2 groups.
func TestMergeSameColorBlocks(t *testing.T) {
	visible := []bool{true, true, true}
	elements := []RenderedElement{
		{Primary: "branch", BgColor: "#008080"}, // teal
		{Primary: "status", BgColor: "#008080"}, // teal — should merge with previous
		{Primary: "time", BgColor: "#ffd700"},   // gold — different, new group
	}
	blocks := mergeSameColorBlocks(elements, visible)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	// First block: teal, indices 0 and 1.
	if blocks[0].bgColor != "#008080" {
		t.Errorf("block[0].bgColor = %q, want %q", blocks[0].bgColor, "#008080")
	}
	if len(blocks[0].indices) != 2 {
		t.Errorf("block[0] should have 2 indices, got %d", len(blocks[0].indices))
	}
	// Second block: gold, index 2.
	if blocks[1].bgColor != "#ffd700" {
		t.Errorf("block[1].bgColor = %q, want %q", blocks[1].bgColor, "#ffd700")
	}
	if len(blocks[1].indices) != 1 {
		t.Errorf("block[1] should have 1 index, got %d", len(blocks[1].indices))
	}
}

// TestMergeSameColorBlocks_NoBg: elements without BgColor are never merged.
func TestMergeSameColorBlocks_NoBg(t *testing.T) {
	visible := []bool{true, true}
	elements := []RenderedElement{
		{Primary: "alpha"}, // no bg
		{Primary: "beta"},  // no bg
	}
	blocks := mergeSameColorBlocks(elements, visible)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks (no merge without bg), got %d", len(blocks))
	}
	for i, b := range blocks {
		if b.bgColor != "" {
			t.Errorf("block[%d].bgColor should be empty, got %q", i, b.bgColor)
		}
		if len(b.indices) != 1 {
			t.Errorf("block[%d] should have exactly 1 index, got %d", i, len(b.indices))
		}
	}
}

// TestMergeSameColorBlocks_HiddenSkipped: hidden element breaks adjacency.
func TestMergeSameColorBlocks_HiddenSkipped(t *testing.T) {
	visible := []bool{true, false, true}
	elements := []RenderedElement{
		{Primary: "a", BgColor: "#008080"},
		{Primary: "b", BgColor: "#008080"}, // hidden — breaks adjacency
		{Primary: "c", BgColor: "#008080"},
	}
	blocks := mergeSameColorBlocks(elements, visible)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks (hidden breaks adjacency), got %d", len(blocks))
	}
	if blocks[0].indices[0] != 0 {
		t.Errorf("block[0] index should be 0, got %d", blocks[0].indices[0])
	}
	if blocks[1].indices[0] != 2 {
		t.Errorf("block[1] index should be 2, got %d", blocks[1].indices[0])
	}
}

// TestPadLineToWidth: short line gets padded to target width.
func TestPadLineToWidth(t *testing.T) {
	line := "hello" // display width 5
	got := PadLineToWidth(line, 10, "#282828")
	stripped := StripANSI(got)
	w := DisplayWidth(stripped)
	if w != 10 {
		t.Errorf("expected display width 10 after pad, got %d (stripped: %q)", w, stripped)
	}
	if !strings.HasPrefix(stripped, "hello") {
		t.Errorf("padded line should start with 'hello', got: %q", stripped)
	}
}

// TestPadLineToWidth_AlreadyFits: no extra padding when line is already wide enough.
func TestPadLineToWidth_AlreadyFits(t *testing.T) {
	line := "hello world" // display width 11
	got := PadLineToWidth(line, 5, "#282828")
	if got != line {
		t.Errorf("line already at/above width should be returned unchanged, got: %q", got)
	}
}

// TestPadLineToWidth_EmptyFillColor: no padding without a fill color.
func TestPadLineToWidth_EmptyFillColor(t *testing.T) {
	line := "hello"
	got := PadLineToWidth(line, 20, "")
	if got != line {
		t.Errorf("empty fill color should return unchanged, got: %q", got)
	}
}

// TestSmartAlign_Basic verifies that multi-line alignment produces the correct
// pad structure to align block boundaries across lines.
func TestSmartAlign_Basic(t *testing.T) {
	// Line 0: [ "ab"(bg) ]  → block ends at position 4 (2 text + 2 padding)
	// Line 1: [ "a"(bg) ]   → block ends at position 3 (1 text + 2 padding)
	// Both have the same bg, so they cluster; line 1 needs 1 pad on element 0.
	line0 := []RenderedElement{
		{Primary: "ab", BgColor: "#008080"},
	}
	line1 := []RenderedElement{
		{Primary: "a", BgColor: "#008080"},
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads

	if len(pads) != 2 {
		t.Fatalf("expected 2 pad slices, got %d", len(pads))
	}
	if len(pads[0]) != 1 || len(pads[1]) != 1 {
		t.Fatalf("expected 1 pad per line, got %d and %d", len(pads[0]), len(pads[1]))
	}

	// Line 0 is already at the max boundary — no pad needed.
	if pads[0][0] != 0 {
		t.Errorf("line 0 element 0: expected pad 0, got %d", pads[0][0])
	}
	// Line 1's block ends 1 short of line 0's boundary — needs pad 1.
	if pads[1][0] != 1 {
		t.Errorf("line 1 element 0: expected pad 1, got %d", pads[1][0])
	}

	// Shrinks should be nil (no shrinkage).
	if result.Shrinks != nil {
		t.Errorf("expected nil Shrinks, got %v", result.Shrinks)
	}
}

// TestSmartAlign_SingleLine: a single line produces all-zero pads.
func TestSmartAlign_SingleLine(t *testing.T) {
	line := []RenderedElement{
		{Primary: "hello", BgColor: "#008080"},
		{Primary: "world", BgColor: "#ffd700"},
	}
	allLines := [][]RenderedElement{line}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads
	if len(pads) != 1 {
		t.Fatalf("expected 1 pad slice, got %d", len(pads))
	}
	for i, p := range pads[0] {
		if p != 0 {
			t.Errorf("single-line pads[0][%d] = %d, want 0", i, p)
		}
	}
}

// TestSmartAlign_NoBgColor: non-block mode aligns separator positions.
// Line 0: "foo"(3) "bar"(3) → boundaries at pos 3 and 6
// Line 1: "x"(1) "yz"(2) → boundaries at pos 1 and 3
// After level-0 alignment: line 1 elem 0 gets +2 (align to pos 3).
// After level-1 alignment: line 1 elem 1 gets +1 (align to pos 6).
// No trailing fill in non-block mode.
func TestSmartAlign_NoBgColor(t *testing.T) {
	line0 := []RenderedElement{
		{Primary: "foo"}, // no bg
		{Primary: "bar"}, // no bg
	}
	line1 := []RenderedElement{
		{Primary: "x"},  // no bg
		{Primary: "yz"}, // no bg
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads
	if len(pads) != 2 {
		t.Fatalf("expected 2 pad slices, got %d", len(pads))
	}
	want := [][]int{{0, 0}, {2, 1}}
	for li, linePads := range pads {
		for ei, p := range linePads {
			if p != want[li][ei] {
				t.Errorf("no-bg pads[%d][%d] = %d, want %d", li, ei, p, want[li][ei])
			}
		}
	}
}

// TestSmartAlign_NoBgColor_WithSeparator verifies non-block alignment with
// a visible separator. Separator width is included in position calculation.
// Line 0: "model"(5) " │ "(3) "main"(4)  → boundaries at 5 and 12
// Line 1: "cwd"(3)   " │ "(3) "v1.2"(4)  → boundaries at 3 and 10
// Level 0: align 3 → 5, pad line 1 elem 0 by 2.
// Level 1: line 1 adjusted = 10 + 2 = 12. Equals line 0 → no pad.
func TestSmartAlign_NoBgColor_WithSeparator(t *testing.T) {
	line0 := []RenderedElement{
		{Primary: "model"},
		{Primary: "main"},
	}
	line1 := []RenderedElement{
		{Primary: "cwd"},
		{Primary: "v1.2"},
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), " │ ", 0)
	want := [][]int{{0, 0}, {2, 0}}
	for li, lp := range result.Pads {
		for ei, p := range lp {
			if p != want[li][ei] {
				t.Errorf("pads[%d][%d] = %d, want %d", li, ei, p, want[li][ei])
			}
		}
	}
}

// TestSmartAlign_NoBgColor_NoTrailingFill verifies non-block mode skips
// trailing fill (trailing spaces invisible without BgColor).
func TestSmartAlign_NoBgColor_NoTrailingFill(t *testing.T) {
	// Line 0: "abcde"(5)  → boundary at 5
	// Line 1: "xy"(2)     → boundary at 2
	// Level 0: align 2 → 5, pad line 1 elem 0 by 3.
	// No trailing fill → line 1 total = 2 + 3 = 5, same as line 0.
	// BUT this is coincidental. The key point: no EXTRA trailing fill pad.
	line0 := []RenderedElement{{Primary: "abcde"}}
	line1 := []RenderedElement{{Primary: "xy"}}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	// Line 0 should get 0 pad (it defines the max).
	if result.Pads[0][0] != 0 {
		t.Errorf("pads[0][0] = %d, want 0", result.Pads[0][0])
	}
	// Line 1 should get exactly 3 (alignment only, no trailing fill).
	if result.Pads[1][0] != 3 {
		t.Errorf("pads[1][0] = %d, want 3", result.Pads[1][0])
	}
}

// TestSmartAlign_NoBgColor_DifferentElementCounts: lines with different
// numbers of elements. Only matching blockIdx levels get aligned.
func TestSmartAlign_NoBgColor_DifferentElementCounts(t *testing.T) {
	// Line 0: "A"(1) "B"(1) "C"(1) → boundaries at 1, 2, 3
	// Line 1: "XX"(2) "Y"(1)        → boundaries at 2, 3
	// Level 0: line 0 pos 1, line 1 pos 2. Align to 2 → pad line 0 elem 0 by 1.
	// Level 1: line 0 adjusted = 2 + 1 = 3. line 1 = 3. Equal → no pad.
	// Level 2: only line 0 has blockIdx=2 → no alignment (need ≥2 lines).
	line0 := []RenderedElement{
		{Primary: "A"}, {Primary: "B"}, {Primary: "C"},
	}
	line1 := []RenderedElement{
		{Primary: "XX"}, {Primary: "Y"},
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	want0 := []int{1, 0, 0}
	for ei, p := range result.Pads[0] {
		if p != want0[ei] {
			t.Errorf("pads[0][%d] = %d, want %d", ei, p, want0[ei])
		}
	}
	for ei, p := range result.Pads[1] {
		if p != 0 {
			t.Errorf("pads[1][%d] = %d, want 0", ei, p)
		}
	}
}

// TestSmartAlign_EqualWidths: lines with equal widths need no alignment.
func TestSmartAlign_EqualWidths(t *testing.T) {
	line0 := []RenderedElement{{Primary: "ab", BgColor: "#111"}}
	line1 := []RenderedElement{{Primary: "cd", BgColor: "#222"}}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	for li, lp := range result.Pads {
		for ei, p := range lp {
			if p != 0 {
				t.Errorf("equal-width pads[%d][%d] = %d, want 0", li, ei, p)
			}
		}
	}
}

// TestSmartAlign_SpreadExceedsMax: boundaries too far apart are not clustered,
// but trailing fill still applies (shorter line padded to match longer).
func TestSmartAlign_SpreadExceedsMax(t *testing.T) {
	line0 := []RenderedElement{{Primary: strings.Repeat("x", 30), BgColor: "#111"}}
	line1 := []RenderedElement{{Primary: "a", BgColor: "#222"}}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads
	// Line 0 is widest — no padding (B constraint: widest line's last group = 0).
	if pads[0][0] != 0 {
		t.Errorf("long line should get no pad, got %d", pads[0][0])
	}
	// Line 1 gets trailing fill: (30+2) - (1+2) = 29.
	if pads[1][0] != 29 {
		t.Errorf("short line should get trailing fill 29, got %d", pads[1][0])
	}
}

// TestSmartAlign_MultiBlock: 3 lines × 3 colored blocks, verifying multi-cluster alignment.
func TestSmartAlign_MultiBlock(t *testing.T) {
	line0 := []RenderedElement{
		{Primary: "model", BgColor: "#a00"},
		{Primary: "cwd-path", BgColor: "#0a0"},
		{Primary: "git", BgColor: "#00a"},
	}
	line1 := []RenderedElement{
		{Primary: "ctx", BgColor: "#a00"},
		{Primary: "tokens", BgColor: "#0a0"},
		{Primary: "cost", BgColor: "#00a"},
	}
	line2 := []RenderedElement{
		{Primary: "usage", BgColor: "#a00"},
		{Primary: "weekly", BgColor: "#0a0"},
		{Primary: "ver", BgColor: "#00a"},
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads

	for li, lp := range pads {
		for ei, p := range lp {
			if p < 0 {
				t.Errorf("pads[%d][%d] = %d, want >= 0", li, ei, p)
			}
		}
	}

	anyPad := false
	for _, lp := range pads {
		for _, p := range lp {
			if p > 0 {
				anyPad = true
			}
		}
	}
	if !anyPad {
		t.Error("expected some alignment padding for blocks with different widths")
	}
}

// TestSmartAlign_VariancePreferred: when two clusters have the same line
// coverage, the algorithm should prefer the one with lower variance.
func TestSmartAlign_VariancePreferred(t *testing.T) {
	line0 := []RenderedElement{
		{Primary: "AAAA", BgColor: "#a00"},
		{Primary: "BB", BgColor: "#0a0"},
	}
	line1 := []RenderedElement{
		{Primary: "AAAAA", BgColor: "#a00"},
		{Primary: "B", BgColor: "#0a0"},
	}
	line2 := []RenderedElement{
		{Primary: "AAAAAA", BgColor: "#a00"},
		{Primary: "BBB", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads

	for li, lp := range pads {
		for ei, p := range lp {
			if p < 0 {
				t.Errorf("pads[%d][%d] = %d, want >= 0", li, ei, p)
			}
		}
	}

	var allPads []int
	for _, lp := range pads {
		allPads = append(allPads, lp...)
	}
	maxPad := 0
	for _, p := range allPads {
		if p > maxPad {
			maxPad = p
		}
	}
	if maxPad > 10 {
		t.Errorf("max pad %d seems excessive for this input, variance minimization may not be working", maxPad)
	}
}

// TestSmartAlign_SeparatorWidth verifies that a non-zero uniform separator
// does not break the algorithm — relative boundary positions are preserved
// and alignment still works. The actual separator-affects-alignment proof
// is in TestSmartAlign_CustomSeparatorAfter (mixed separator widths).
func TestSmartAlign_SeparatorWidth(t *testing.T) {
	// With uniform separator, relative first-boundary positions are unchanged,
	// but second boundaries shift by the same amount. The algorithm should
	// still produce valid, non-negative pads.
	line0 := []RenderedElement{
		{Primary: "AAA", BgColor: "#a00"},
		{Primary: "B", BgColor: "#0a0"},
	}
	line1 := []RenderedElement{
		{Primary: "AA", BgColor: "#a00"},
		{Primary: "BB", BgColor: "#0a0"},
	}
	line2 := []RenderedElement{
		{Primary: "A", BgColor: "#a00"},
		{Primary: "BBB", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	result := SmartAlign(allLines, testElemWidths(allLines), " | ", 0)
	pads := result.Pads

	for li, lp := range pads {
		for ei, p := range lp {
			if p < 0 {
				t.Errorf("pads[%d][%d] = %d, want >= 0", li, ei, p)
			}
		}
	}
}

// TestSmartAlign_CustomSeparatorAfter verifies that per-element SeparatorAfter
// overrides are used for position calculation.
func TestSmartAlign_CustomSeparatorAfter(t *testing.T) {
	// Line 0: [AA](4) customSep(1) [B](3)  → boundary at 4, then 4+1+3=8
	// Line 1: [AA](4) defaultSep(3) [B](3)  → boundary at 4, then 4+3+3=10
	// Second boundaries differ by 2 → line 0 gets +2 pad on second element.
	line0 := []RenderedElement{
		{Primary: "AA", BgColor: "#a00", SeparatorAfter: " "},
		{Primary: "B", BgColor: "#0a0"},
	}
	line1 := []RenderedElement{
		{Primary: "AA", BgColor: "#a00"},
		{Primary: "B", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), " | ", 0)
	pads := result.Pads

	// First boundaries are equal (both at 4) → no padding on first element.
	if pads[0][0] != 0 {
		t.Errorf("line 0 elem 0: expected 0, got %d", pads[0][0])
	}
	if pads[1][0] != 0 {
		t.Errorf("line 1 elem 0: expected 0, got %d", pads[1][0])
	}
	// Second boundary: line 0 at 8, line 1 at 10. Line 0 needs +2.
	if pads[0][1] != 2 {
		t.Errorf("line 0 elem 1: expected 2, got %d", pads[0][1])
	}
	if pads[1][1] != 0 {
		t.Errorf("line 1 elem 1: expected 0, got %d", pads[1][1])
	}
}

// TestSmartAlign_TrailingFill verifies that all lines are padded to the same
// total width via trailing fill on the last colored group.
func TestSmartAlign_TrailingFill(t *testing.T) {
	// Line 0: [ABCDE](7)  → width 7
	// Line 1: [AB](4)     → width 4
	// Line 2: [ABC](5)    → width 5
	// No alignment possible (single block per line, spread=3≤16 → cluster).
	// After alignment: all align to 7. Trailing fill: line1 +3, line2 +2.
	line0 := []RenderedElement{{Primary: "ABCDE", BgColor: "#a00"}}
	line1 := []RenderedElement{{Primary: "AB", BgColor: "#a00"}}
	line2 := []RenderedElement{{Primary: "ABC", BgColor: "#a00"}}
	allLines := [][]RenderedElement{line0, line1, line2}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads

	// Compute effective widths: elemWidth + pad should be equal across all lines.
	ew := testElemWidths(allLines)
	widths := make([]int, 3)
	for li := range allLines {
		widths[li] = ew[li][0] + pads[li][0]
	}
	if widths[0] != widths[1] || widths[1] != widths[2] {
		t.Errorf("lines should have equal total width, got %d, %d, %d", widths[0], widths[1], widths[2])
	}
}

// TestSmartAlign_BConstraint verifies the widest line's last group gets zero
// extra padding (neither alignment nor trailing fill).
func TestSmartAlign_BConstraint(t *testing.T) {
	// Line 0: [ABCDEF](8) [XY](4) → widths 8+4=12
	// Line 1: [AB](4)     [X](3)  → widths 4+3=7
	// Line 0 is widest. Its last element (index 1) should get 0 padding.
	line0 := []RenderedElement{
		{Primary: "ABCDEF", BgColor: "#a00"},
		{Primary: "XY", BgColor: "#0a0"},
	}
	line1 := []RenderedElement{
		{Primary: "AB", BgColor: "#a00"},
		{Primary: "X", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads

	// Line 0 is widest — its last element should have 0 pad.
	if pads[0][1] != 0 {
		t.Errorf("widest line's last element should have 0 pad, got %d", pads[0][1])
	}

	// Line 1 should get trailing fill to match line 0's total width.
	ew := testElemWidths(allLines)
	line0Width := ew[0][0] + pads[0][0] + ew[0][1] + pads[0][1]
	line1Width := ew[1][0] + pads[1][0] + ew[1][1] + pads[1][1]
	if line0Width != line1Width {
		t.Errorf("lines should have equal total width: line0=%d, line1=%d", line0Width, line1Width)
	}
}

// TestSmartAlign_TrailingFill_EqualWidths: equal-width lines get zero trailing fill.
func TestSmartAlign_TrailingFill_EqualWidths(t *testing.T) {
	line0 := []RenderedElement{{Primary: "AB", BgColor: "#111"}}
	line1 := []RenderedElement{{Primary: "CD", BgColor: "#222"}}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	for li, lp := range result.Pads {
		for ei, p := range lp {
			if p != 0 {
				t.Errorf("equal-width trailing fill pads[%d][%d] = %d, want 0", li, ei, p)
			}
		}
	}
}

// TestSmartAlign_TrailingFill_NoBgLines: lines without colored blocks get no trailing fill.
func TestSmartAlign_TrailingFill_NoBgLines(t *testing.T) {
	line0 := []RenderedElement{{Primary: "ABCDE", BgColor: "#a00"}}
	line1 := []RenderedElement{{Primary: "X"}} // no bg
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)

	// Line 1 has no colored blocks → no trailing fill.
	for ei, p := range result.Pads[1] {
		if p != 0 {
			t.Errorf("no-bg line pads[1][%d] = %d, want 0", ei, p)
		}
	}
}

// TestSmartAlign_WidthAware_NoShrinkNeeded: alignment fits within maxWidth.
func TestSmartAlign_WidthAware_NoShrinkNeeded(t *testing.T) {
	line0 := []RenderedElement{{Primary: "AB", BgColor: "#a00"}}
	line1 := []RenderedElement{{Primary: "A", BgColor: "#a00"}}
	allLines := [][]RenderedElement{line0, line1}
	// maxWidth=10, line widths are 4 and 3 — plenty of room.
	result := SmartAlign(allLines, testElemWidths(allLines), "", 10)
	if result.Shrinks != nil {
		t.Errorf("expected nil Shrinks when alignment fits, got %v", result.Shrinks)
	}
}

// TestSmartAlign_WidthAware_ShrinkNeeded: alignment causes slight overflow,
// variable-width element is shrunk to fit within maxWidth.
func TestSmartAlign_WidthAware_ShrinkNeeded(t *testing.T) {
	// Line 0: [variable:7chars](9, var, min=5) [BB](4) → total 13
	// Line 1: [short:5chars](7)                [BBB](5) → total 12
	// Boundaries: {9,13} and {7,12}. Cluster target=13, line 1 pad=1 (≤maxGroupPad).
	// After padding: maxEff=13 > maxWidth=12 → shrink line 0's variable by 1.
	line0 := []RenderedElement{
		{Primary: "xxxxxxx", BgColor: "#a00", IsVariable: true, MinWidth: 5},
		{Primary: "BB", BgColor: "#0a0"},
	}
	line1 := []RenderedElement{
		{Primary: "short", BgColor: "#a00"},
		{Primary: "BBB", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line0, line1}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 12)

	// Should have shrinkage on line 0's variable element.
	if result.Shrinks == nil {
		t.Fatal("expected Shrinks to be non-nil")
	}
	if result.Shrinks[0][0] <= 0 {
		t.Errorf("line 0 elem 0 should be shrunk, got shrink=%d", result.Shrinks[0][0])
	}
	// Shrinkage should not exceed capacity.
	capacity := ew[0][0] - line0[0].MinWidth
	if result.Shrinks[0][0] > capacity {
		t.Errorf("shrink %d exceeds capacity %d", result.Shrinks[0][0], capacity)
	}
}

// TestSmartAlign_WidthAware_Unlimited: maxWidth=0 means no width checking.
func TestSmartAlign_WidthAware_Unlimited(t *testing.T) {
	line0 := []RenderedElement{
		{Primary: strings.Repeat("x", 50), BgColor: "#a00", IsVariable: true, MinWidth: 10},
	}
	line1 := []RenderedElement{
		{Primary: "A", BgColor: "#a00"},
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	if result.Shrinks != nil {
		t.Errorf("maxWidth=0 should not cause shrinkage, got %v", result.Shrinks)
	}
}

// TestSmartAlign_WidthAware_InfeasibleCluster: overflow exceeds all shrink
// capacity → cluster skipped, no alignment rather than over-shrink.
func TestSmartAlign_WidthAware_InfeasibleCluster(t *testing.T) {
	// Line 0: [very-wide](32) → boundary at 32
	// Line 1: [tiny](5)       → boundary at 5
	// maxWidth=15. Even fully shrinking line 0 (cap=32-10=22) would give 10,
	// still fits. But alignment wants to pad line 1 to 32 → 32 > 15.
	// Let's make it truly infeasible: no variable-width elements.
	line0 := []RenderedElement{{Primary: strings.Repeat("x", 30), BgColor: "#a00"}}
	line1 := []RenderedElement{{Primary: "A", BgColor: "#a00"}}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 15)
	// Spread is 32-3=29 > maxAlignSpread(16), so no clustering occurs.
	// Trailing fill capped at maxWidth. No shrinkage possible (no variable elements).
	if result.Shrinks != nil {
		t.Errorf("expected nil Shrinks (no variable elements), got %v", result.Shrinks)
	}
}

// TestSmartAlign_WidthAware_AlreadyAtMinWidth: variable-width element already
// at MinWidth has zero remaining capacity.
func TestSmartAlign_WidthAware_AlreadyAtMinWidth(t *testing.T) {
	// Simulate post-budget: elemWidth == MinWidth → 0 capacity.
	line0 := []RenderedElement{
		{Primary: "ABCDEFGH", BgColor: "#a00", IsVariable: true, MinWidth: 10},
	}
	line1 := []RenderedElement{
		{Primary: "A", BgColor: "#a00"},
	}
	allLines := [][]RenderedElement{line0, line1}
	// Override elemWidths to simulate budgetLine already shrunk to MinWidth.
	ew := [][]int{{10}, {3}}
	result := SmartAlign(allLines, ew, "", 8)
	// Can't shrink below MinWidth. Trailing fill capped at maxWidth.
	if result.Shrinks != nil {
		t.Errorf("expected nil Shrinks (already at MinWidth), got %v", result.Shrinks)
	}
}

// TestSmartAlign_DisplacementPenalty: a cluster needing large per-line padding
// (> maxGroupPad) gets its effective coverage demoted, allowing a less aggressive
// alignment to win.
func TestSmartAlign_DisplacementPenalty(t *testing.T) {
	// Line 0: [A](3)                     [BBBB](6) → boundaries at 3 and 9
	// Line 1: [AAAAAAAAAAAA](14)         [B](3)    → boundaries at 14 and 17
	// First boundary cluster {3, 14}: spread=11≤16, target=14, line0 pad=11.
	// maxLinePad=11 > maxGroupPad(6) → effectiveLines demoted from 2 to 1 → skipped.
	// Result: first boundaries NOT aligned (displacement too high).
	line0 := []RenderedElement{
		{Primary: "A", BgColor: "#a00"},
		{Primary: "BBBB", BgColor: "#0a0"},
	}
	line1 := []RenderedElement{
		{Primary: "AAAAAAAAAAAA", BgColor: "#a00"},
		{Primary: "B", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line0, line1}
	result := SmartAlign(allLines, testElemWidths(allLines), "", 0)
	pads := result.Pads

	// Line 0's first element should NOT get 11 columns of padding.
	if pads[0][0] > maxGroupPad {
		t.Errorf("displacement penalty failed: line 0 elem 0 pad=%d, want ≤%d", pads[0][0], maxGroupPad)
	}
}

// TestSmartAlign_CrossBlockBoundaryContamination tests that boundaries from
// different block indices on the same line do not contaminate each other's
// alignment cluster. When a later block's boundary (e.g. G3) falls within the
// position window of an earlier block's cluster (G2), the "lastPosPerLine"
// logic shadows the earlier boundary, leaving it assigned but un-padded.
//
// Scenario: 3 lines × 3 colored blocks. Line 2's G3 boundary (original
// position=16) falls within maxAlignSpread(16) of the G2 boundaries
// (positions 13–19), contaminating the G2 cluster.
func TestSmartAlign_CrossBlockBoundaryContamination(t *testing.T) {
	// Line 0: [AAAA](6)  [BBBBBBB](9)  [CCCCCCCCCCCCCCCCCCCC](22)
	//         G1 pos=6     G2 pos=15     G3 pos=37
	// Line 1: [AAAAAA](8) [BBBBBBBBB](11) [CC](4)
	//         G1 pos=8     G2 pos=19     G3 pos=23
	// Line 2: [AAAA](6)   [BBBBB](7)    [C](3)
	//         G1 pos=6     G2 pos=13     G3 pos=16
	//
	// After G1 alignment (target=8): padOffset=[2, 0, 2]
	// Adjusted G2 positions: L0=17, L1=19, L2=15
	// G2 spread=4, max needed pad=4 ≤ maxGroupPad(6) → feasible 3-line cluster.
	//
	// Bug: L2G3 (original pos=16) sits in the G2 window. lastPosPerLine picks
	// L2G3 over L2G2, so L2G2 gets assigned without padding.
	line0 := []RenderedElement{
		{Primary: "AAAA", BgColor: "#a00"},
		{Primary: "BBBBBBB", BgColor: "#0a0"},
		{Primary: "CCCCCCCCCCCCCCCCCCCC", BgColor: "#00a"},
	}
	line1 := []RenderedElement{
		{Primary: "AAAAAA", BgColor: "#a00"},
		{Primary: "BBBBBBBBB", BgColor: "#0a0"},
		{Primary: "CC", BgColor: "#00a"},
	}
	line2 := []RenderedElement{
		{Primary: "AAAA", BgColor: "#a00"},
		{Primary: "BBBBB", BgColor: "#0a0"},
		{Primary: "C", BgColor: "#00a"},
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	pads := result.Pads

	// Compute effective G2 boundary position for each line:
	// = sum of (elemWidth + pad) for elements 0 and 1.
	g2Boundaries := make([]int, 3)
	for li := range allLines {
		g2Boundaries[li] = ew[li][0] + pads[li][0] + ew[li][1] + pads[li][1]
	}

	// All G2 boundaries must be aligned to the same column (19).
	// The correct alignment: pads[0][1]=2, pads[1][1]=0, pads[2][1]=4.
	if g2Boundaries[0] != g2Boundaries[1] || g2Boundaries[1] != g2Boundaries[2] {
		t.Errorf("G2 boundaries not aligned: L0=%d, L1=%d, L2=%d (want all equal to 19)",
			g2Boundaries[0], g2Boundaries[1], g2Boundaries[2])
	}
}

// TestSmartAlign_ShrinkOffsetGranularity verifies that shrinkage on a later
// block (G3) does not corrupt the adjusted position of an earlier block (G2).
// The per-line shrinkOffset total includes G3 shrinkage, but G2's visual
// position is unaffected by G3 shrinkage — only shrinkage BEFORE the boundary
// should be subtracted.
func TestSmartAlign_ShrinkOffsetGranularity(t *testing.T) {
	// Line 0: [AAAAA](7) [BBBBBB](8) [var:CCCCCCCCCCCCCCCCCCCC](22,min=5)
	//         G1=7  G2=15  G3=37  lineWidth=37
	// Line 1: [AAAAAAAA](10) [BBB](5) [CCC](5)
	//         G1=10  G2=15  G3=20  lineWidth=20
	// maxWidth=25.
	//
	// Level 0: target=10, L0 needs +3. L0 eff=40>25 → shrink G3 by 15.
	// Level 1: G2 adjusted positions should be L0=15+3-0=18, L1=15.
	//   Bug: uses per-line shrinkOffset → L0=15+3-15=3, G2 cluster skipped.
	//   Fix: only count shrinks before G2 → shrinkUpTo=0, L0=18.
	line0 := []RenderedElement{
		{Primary: "AAAAA", BgColor: "#a00"},
		{Primary: "BBBBBB", BgColor: "#0a0"},
		{Primary: "CCCCCCCCCCCCCCCCCCCC", BgColor: "#00a", IsVariable: true, MinWidth: 5},
	}
	line1 := []RenderedElement{
		{Primary: "AAAAAAAA", BgColor: "#a00"},
		{Primary: "BBB", BgColor: "#0a0"},
		{Primary: "CCC", BgColor: "#00a"},
	}
	allLines := [][]RenderedElement{line0, line1}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 25)
	pads := result.Pads

	// G2 boundaries should be aligned at 18.
	g2L0 := ew[0][0] + pads[0][0] + ew[0][1] + pads[0][1]
	g2L1 := ew[1][0] + pads[1][0] + ew[1][1] + pads[1][1]
	if g2L0 != g2L1 {
		t.Errorf("G2 boundaries not aligned: L0=%d, L1=%d (want both 18)", g2L0, g2L1)
	}
}

// TestSmartAlign_ShrinkAllOverwidthLines verifies that when multiple lines
// exceed maxWidth by the same amount, ALL are shrunk — not just the first.
func TestSmartAlign_ShrinkAllOverwidthLines(t *testing.T) {
	// L0 and L1 identical, L2 wider.
	// L0: [var:AAAA](6,min=3) [BBBBBBBBBB](12) → lineWidth=18
	// L1: [var:AAAA](6,min=3) [BBBBBBBBBB](12) → lineWidth=18
	// L2: [AAAAAAAA](10)      [BBBBBBBB](10)   → lineWidth=20
	// maxWidth=21.
	//
	// Level 0: G1 at {6,6,10}. Target=10. L0+4, L1+4. padOffset=[4,4,0].
	// L0 eff=22>21, L1 eff=22>21. Both at maxEff=22.
	// Bug: distributeShrinkage overflow=1, shrinks only L0. L1 stays at 22.
	// Fix: per-line shrinkage, both L0 and L1 shrunk by 1.
	line0 := []RenderedElement{
		{Primary: "AAAA", BgColor: "#a00", IsVariable: true, MinWidth: 3},
		{Primary: "BBBBBBBBBB", BgColor: "#0a0"},
	}
	line1 := []RenderedElement{
		{Primary: "AAAA", BgColor: "#a00", IsVariable: true, MinWidth: 3},
		{Primary: "BBBBBBBBBB", BgColor: "#0a0"},
	}
	line2 := []RenderedElement{
		{Primary: "AAAAAAAA", BgColor: "#a00"},
		{Primary: "BBBBBBBB", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	ew := testElemWidths(allLines)
	maxWidth := 21
	result := SmartAlign(allLines, ew, "", maxWidth)

	// No line's effective width (including alignment padding, shrinkage,
	// and trailing fill) should exceed maxWidth.
	for li := range allLines {
		w := 0
		for ei := range allLines[li] {
			s := 0
			if result.Shrinks != nil && li < len(result.Shrinks) {
				s = result.Shrinks[li][ei]
			}
			w += ew[li][ei] - s + result.Pads[li][ei]
		}
		if w > maxWidth {
			t.Errorf("line %d effective width %d > maxWidth %d", li, w, maxWidth)
		}
	}
}

// ---------------------------------------------------------------------------
// Black-box invariant tests: verify SmartAlign output properties without
// relying on internal algorithm details.
// ---------------------------------------------------------------------------

// smartAlignInvariants runs a battery of invariant checks on a SmartAlign
// result. It returns all violations as error strings.
func smartAlignInvariants(
	t *testing.T,
	allLines [][]RenderedElement,
	ew [][]int,
	sep string,
	maxWidth int,
	result AlignResult,
) {
	t.Helper()

	numLines := len(allLines)

	// --- Invariant 1: Non-negative pads ---
	for li := 0; li < numLines; li++ {
		for ei, p := range result.Pads[li] {
			if p < 0 {
				t.Errorf("I1: pads[%d][%d] = %d < 0", li, ei, p)
			}
		}
	}

	// --- Invariant 2: Non-negative shrinks, within capacity ---
	if result.Shrinks != nil {
		for li := 0; li < numLines; li++ {
			for ei, s := range result.Shrinks[li] {
				if s < 0 {
					t.Errorf("I2: shrinks[%d][%d] = %d < 0", li, ei, s)
				}
				e := allLines[li][ei]
				if e.IsVariable && e.MinWidth > 0 {
					cap := ew[li][ei] - e.MinWidth
					if cap < 0 {
						cap = 0
					}
					if s > cap {
						t.Errorf("I2: shrinks[%d][%d] = %d > capacity %d", li, ei, s, cap)
					}
				}
			}
		}
	}

	// Helper: compute effective line width including separators.
	effectiveLineWidth := func(li int) int {
		vis := make([]bool, len(allLines[li]))
		for i := range vis {
			vis[i] = true
		}
		blocks := mergeSameColorBlocks(allLines[li], vis)
		w := 0
		for bi, blk := range blocks {
			if bi > 0 {
				prevBlk := blocks[bi-1]
				lastPrev := prevBlk.indices[len(prevBlk.indices)-1]
				s := sep
				if allLines[li][lastPrev].SeparatorAfter != "" {
					s = allLines[li][lastPrev].SeparatorAfter
				}
				w += DisplayWidth(s)
			}
			for _, idx := range blk.indices {
				shrink := 0
				if result.Shrinks != nil {
					shrink = result.Shrinks[li][idx]
				}
				w += ew[li][idx] - shrink + result.Pads[li][idx]
			}
		}
		return w
	}

	// --- Invariant 3: Width constraint ---
	if maxWidth > 0 {
		for li := 0; li < numLines; li++ {
			w := effectiveLineWidth(li)
			if w > maxWidth {
				t.Errorf("I3: line %d effective width %d > maxWidth %d", li, w, maxWidth)
			}
		}
	}

	// --- Invariant 4: Boundary alignment (same blockIdx → same column) ---
	// Only checked when maxWidth=0 (unlimited). Under width constraints,
	// alignment may be infeasible due to overflow / insufficient shrink capacity.
	if maxWidth == 0 {
		type bndry struct {
			lineIdx, blockIdx, lastElemIdx, position int
		}
		var boundaries []bndry
		for li, line := range allLines {
			allVis := make([]bool, len(line))
			for i := range allVis {
				allVis[i] = true
			}
			blocks := mergeSameColorBlocks(line, allVis)
			pos := 0
			coloredIdx := 0
			for bi, blk := range blocks {
				if bi > 0 {
					prevBlk := blocks[bi-1]
					lastPrev := prevBlk.indices[len(prevBlk.indices)-1]
					s := sep
					if line[lastPrev].SeparatorAfter != "" {
						s = line[lastPrev].SeparatorAfter
					}
					pos += DisplayWidth(s)
				}
				bw := 0
				for _, idx := range blk.indices {
					bw += ew[li][idx]
				}
				pos += bw
				if blk.bgColor != "" {
					boundaries = append(boundaries, bndry{li, coloredIdx, blk.indices[len(blk.indices)-1], pos})
					coloredIdx++
				}
			}
		}

		type effBndry struct {
			lineIdx, blockIdx, effPos int
		}
		var effBounds []effBndry
		for _, b := range boundaries {
			padBefore, shrinkBefore := 0, 0
			for ei := 0; ei <= b.lastElemIdx; ei++ {
				padBefore += result.Pads[b.lineIdx][ei]
				if result.Shrinks != nil {
					shrinkBefore += result.Shrinks[b.lineIdx][ei]
				}
			}
			effBounds = append(effBounds, effBndry{b.lineIdx, b.blockIdx, b.position + padBefore - shrinkBefore})
		}

		byBlock := map[int][]effBndry{}
		for _, eb := range effBounds {
			byBlock[eb.blockIdx] = append(byBlock[eb.blockIdx], eb)
		}
		for bi, ebs := range byBlock {
			if len(ebs) < 2 {
				continue
			}
			target := ebs[0].effPos
			allSame := true
			for _, eb := range ebs[1:] {
				if eb.effPos != target {
					allSame = false
					break
				}
			}
			if !allSame {
				minPos, maxPos := ebs[0].effPos, ebs[0].effPos
				for _, eb := range ebs[1:] {
					if eb.effPos < minPos {
						minPos = eb.effPos
					}
					if eb.effPos > maxPos {
						maxPos = eb.effPos
					}
				}
				spread := maxPos - minPos
				if spread <= maxGroupPad {
					t.Errorf("I4: blockIdx %d boundaries not aligned (spread %d ≤ maxGroupPad %d): %v",
						bi, spread, maxGroupPad, ebs)
				}
			}
		}
	}

	// --- Invariant 5: Trailing fill uniformity ---
	// All lines with colored blocks should have equal effective total width.
	var coloredLineWidths []int
	for li := 0; li < numLines; li++ {
		hasColor := false
		for _, e := range allLines[li] {
			if e.BgColor != "" {
				hasColor = true
				break
			}
		}
		if !hasColor {
			continue
		}
		coloredLineWidths = append(coloredLineWidths, effectiveLineWidth(li))
	}
	if len(coloredLineWidths) > 1 {
		target := coloredLineWidths[0]
		for i, w := range coloredLineWidths[1:] {
			if w != target {
				t.Errorf("I5: trailing fill not uniform: line 0 width=%d, line %d width=%d",
					target, i+1, w)
			}
		}
	}
}

func TestSmartAlign_BlackBox_UniformGroups(t *testing.T) {
	// 3 lines × 3 groups, moderate width differences.
	allLines := [][]RenderedElement{
		{
			{Primary: "Opus 4.6", BgColor: "#c678dd"},
			{Primary: "~/project", BgColor: "#3c404d"},
			{Primary: "main", BgColor: "#98c379"},
		},
		{
			{Primary: "78%", BgColor: "#283d60"},
			{Primary: "12.3k", BgColor: "#1e4545"},
			{Primary: "$0.42", BgColor: "#4a3818"},
		},
		{
			{Primary: "5h: 82%", BgColor: "#4a3020"},
			{Primary: "Wk: $4.20", BgColor: "#1a3048"},
			{Primary: "v2.1", BgColor: "#2e3038"},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, " ", 0)
	smartAlignInvariants(t, allLines, ew, " ", 0, result)
}

func TestSmartAlign_BlackBox_NarrowTerminal(t *testing.T) {
	// Same layout but constrained width forces shrinkage.
	allLines := [][]RenderedElement{
		{
			{Primary: "Opus 4.6", BgColor: "#c678dd"},
			{Primary: "~/my/long/project/path", BgColor: "#3c404d", IsVariable: true, MinWidth: 8},
			{Primary: "feature/long-branch", BgColor: "#98c379", IsVariable: true, MinWidth: 6},
		},
		{
			{Primary: "78%", BgColor: "#283d60"},
			{Primary: "12.3k", BgColor: "#1e4545"},
			{Primary: "$0.42", BgColor: "#4a3818"},
		},
		{
			{Primary: "5h: 82%", BgColor: "#4a3020"},
			{Primary: "Wk: $4.20", BgColor: "#1a3048"},
			{Primary: "v2.1", BgColor: "#2e3038"},
		},
	}
	ew := testElemWidths(allLines)
	for _, width := range []int{50, 40, 35, 30} {
		t.Run(strings.Repeat("w", 0)+string(rune('0'+width/10))+string(rune('0'+width%10)), func(t *testing.T) {
			result := SmartAlign(allLines, ew, " ", width)
			smartAlignInvariants(t, allLines, ew, " ", width, result)
		})
	}
}

func TestSmartAlign_BlackBox_AsymmetricGroupCounts(t *testing.T) {
	// Line 0 has 3 groups, Line 1 has 2 groups.
	allLines := [][]RenderedElement{
		{
			{Primary: "model", BgColor: "#a00"},
			{Primary: "path", BgColor: "#0a0"},
			{Primary: "branch", BgColor: "#00a"},
		},
		{
			{Primary: "context", BgColor: "#a00"},
			{Primary: "cost", BgColor: "#0a0"},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, " ", 0)
	smartAlignInvariants(t, allLines, ew, " ", 0, result)
}

func TestSmartAlign_BlackBox_LargeSpread(t *testing.T) {
	// G1 boundaries far apart (>maxAlignSpread) → no G1 alignment.
	allLines := [][]RenderedElement{
		{
			{Primary: "A", BgColor: "#a00"},
			{Primary: "B", BgColor: "#0a0"},
		},
		{
			{Primary: strings.Repeat("X", 20), BgColor: "#a00"},
			{Primary: "Y", BgColor: "#0a0"},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)
}

func TestSmartAlign_BlackBox_AllVariableWidth(t *testing.T) {
	// Every group is variable-width under a tight maxWidth.
	allLines := [][]RenderedElement{
		{
			{Primary: strings.Repeat("A", 15), BgColor: "#a00", IsVariable: true, MinWidth: 5},
			{Primary: strings.Repeat("B", 15), BgColor: "#0a0", IsVariable: true, MinWidth: 5},
		},
		{
			{Primary: strings.Repeat("C", 10), BgColor: "#a00", IsVariable: true, MinWidth: 5},
			{Primary: strings.Repeat("D", 10), BgColor: "#0a0", IsVariable: true, MinWidth: 5},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 20)
	smartAlignInvariants(t, allLines, ew, "", 20, result)
}

func TestSmartAlign_BlackBox_MixedBgAndNoBg(t *testing.T) {
	// Some elements have bg, some don't. Non-bg elements should not
	// receive alignment padding or interfere with colored block alignment.
	allLines := [][]RenderedElement{
		{
			{Primary: "model", BgColor: "#a00"},
			{Primary: " > "},
			{Primary: "path", BgColor: "#0a0"},
		},
		{
			{Primary: "ctx", BgColor: "#a00"},
			{Primary: " > "},
			{Primary: "tokens", BgColor: "#0a0"},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	// Non-bg elements should get zero padding.
	if result.Pads[0][1] != 0 {
		t.Errorf("non-bg element pads[0][1] = %d, want 0", result.Pads[0][1])
	}
	if result.Pads[1][1] != 0 {
		t.Errorf("non-bg element pads[1][1] = %d, want 0", result.Pads[1][1])
	}
}

func TestSmartAlign_BlackBox_IdenticalLines(t *testing.T) {
	// All lines identical → zero padding, zero shrink.
	line := []RenderedElement{
		{Primary: "AAA", BgColor: "#a00"},
		{Primary: "BBB", BgColor: "#0a0"},
	}
	allLines := [][]RenderedElement{line, line, line}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	for li, lp := range result.Pads {
		for ei, p := range lp {
			if p != 0 {
				t.Errorf("identical lines: pads[%d][%d] = %d, want 0", li, ei, p)
			}
		}
	}
}

func TestSmartAlign_BlackBox_ShrinkWithAlignment(t *testing.T) {
	// G1 alignment causes overflow, G3 (variable) gets shrunk, then G2
	// should still align correctly using per-boundary shrink tracking.
	allLines := [][]RenderedElement{
		{
			{Primary: "AA", BgColor: "#a00"},
			{Primary: "BBBBB", BgColor: "#0a0"},
			{Primary: strings.Repeat("C", 20), BgColor: "#00a", IsVariable: true, MinWidth: 5},
		},
		{
			{Primary: "AAAAAAA", BgColor: "#a00"},
			{Primary: "BB", BgColor: "#0a0"},
			{Primary: "CC", BgColor: "#00a"},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 25)
	smartAlignInvariants(t, allLines, ew, "", 25, result)
}

func TestSmartAlign_BlackBox_CustomSeparators(t *testing.T) {
	// Mixed separator widths across lines.
	allLines := [][]RenderedElement{
		{
			{Primary: "AA", BgColor: "#a00", SeparatorAfter: " "},
			{Primary: "BB", BgColor: "#0a0"},
		},
		{
			{Primary: "AA", BgColor: "#a00", SeparatorAfter: " | "},
			{Primary: "BB", BgColor: "#0a0"},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, " ", 0)
	smartAlignInvariants(t, allLines, ew, " ", 0, result)
}

func TestSmartAlign_BlackBox_ManyLinesNarrow(t *testing.T) {
	// Stress: 3 lines, 3 variable groups each, very tight width.
	allLines := [][]RenderedElement{
		{
			{Primary: strings.Repeat("A", 10), BgColor: "#a00", IsVariable: true, MinWidth: 3},
			{Primary: strings.Repeat("B", 12), BgColor: "#0a0", IsVariable: true, MinWidth: 3},
			{Primary: strings.Repeat("C", 8), BgColor: "#00a", IsVariable: true, MinWidth: 3},
		},
		{
			{Primary: strings.Repeat("D", 8), BgColor: "#a00", IsVariable: true, MinWidth: 3},
			{Primary: strings.Repeat("E", 10), BgColor: "#0a0", IsVariable: true, MinWidth: 3},
			{Primary: strings.Repeat("F", 12), BgColor: "#00a", IsVariable: true, MinWidth: 3},
		},
		{
			{Primary: strings.Repeat("G", 12), BgColor: "#a00", IsVariable: true, MinWidth: 3},
			{Primary: strings.Repeat("H", 8), BgColor: "#0a0", IsVariable: true, MinWidth: 3},
			{Primary: strings.Repeat("I", 10), BgColor: "#00a", IsVariable: true, MinWidth: 3},
		},
	}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 25)
	smartAlignInvariants(t, allLines, ew, "", 25, result)
}

// ---------------------------------------------------------------------------
// Boundary alignment edge-case tests: maxGroupPad, maxAlignSpread thresholds,
// displacement penalty, and multi-level interaction.
// ---------------------------------------------------------------------------

// g1Boundary returns the effective G1 boundary position (sum of elem 0
// width + pad) for a line with ≥1 element.
func g1Boundary(ew [][]int, pads [][]int, li int) int {
	return ew[li][0] + pads[li][0]
}

// TestSmartAlign_Edge_SpreadExactlyMaxGroupPad verifies that 2-line G1
// alignment succeeds when spread equals maxGroupPad (6) exactly.
func TestSmartAlign_Edge_SpreadExactlyMaxGroupPad(t *testing.T) {
	// G1 positions: 4 and 10 (spread=6=maxGroupPad).
	// maxLinePad=6 → no penalty → effectiveLines=2 → aligns.
	line0 := []RenderedElement{
		{Primary: "XX", BgColor: "#a00"}, // width 4
		{Primary: "YY", BgColor: "#0a0"}, // width 4
	}
	line1 := []RenderedElement{
		{Primary: "XXXXXXXX", BgColor: "#a00"}, // width 10
		{Primary: "Y", BgColor: "#0a0"},        // width 3
	}
	allLines := [][]RenderedElement{line0, line1}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	b0 := g1Boundary(ew, result.Pads, 0)
	b1 := g1Boundary(ew, result.Pads, 1)
	if b0 != b1 {
		t.Errorf("spread=6: G1 should align, got L0=%d L1=%d", b0, b1)
	}
}

// TestSmartAlign_Edge_SpreadExceedsMaxGroupPad verifies that 2-line G1
// alignment is skipped when spread is maxGroupPad+1 (7).
func TestSmartAlign_Edge_SpreadExceedsMaxGroupPad(t *testing.T) {
	// G1 positions: 4 and 11 (spread=7>maxGroupPad).
	// effectiveLines=2-1=1<2 → cluster skipped.
	line0 := []RenderedElement{
		{Primary: "XX", BgColor: "#a00"}, // width 4
		{Primary: "YY", BgColor: "#0a0"}, // width 4
	}
	line1 := []RenderedElement{
		{Primary: "XXXXXXXXX", BgColor: "#a00"}, // width 11
		{Primary: "Y", BgColor: "#0a0"},         // width 3
	}
	allLines := [][]RenderedElement{line0, line1}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	b0 := g1Boundary(ew, result.Pads, 0)
	b1 := g1Boundary(ew, result.Pads, 1)
	if b0 == b1 {
		t.Errorf("spread=7: G1 should NOT align for 2 lines, but both at %d", b0)
	}
	// L0 should keep its natural position (no alignment pad on G1).
	if result.Pads[0][0] != 0 {
		t.Errorf("spread=7: L0 G1 pad should be 0, got %d", result.Pads[0][0])
	}
}

// TestSmartAlign_Edge_ThreeLinePenaltyStillAligns verifies that with 3 lines,
// a displacement penalty (effectiveLines=3→2) still allows alignment because
// 2 ≥ 2.
func TestSmartAlign_Edge_ThreeLinePenaltyStillAligns(t *testing.T) {
	// G1 positions: 4, 7, 11 (maxLinePad=7>maxGroupPad).
	// effectiveLines=3-1=2 ≥ 2 → cluster still applied to all 3 lines.
	line0 := []RenderedElement{
		{Primary: "XX", BgColor: "#a00"}, // width 4
		{Primary: "YY", BgColor: "#0a0"}, // width 4
	}
	line1 := []RenderedElement{
		{Primary: "XXXXX", BgColor: "#a00"}, // width 7
		{Primary: "YY", BgColor: "#0a0"},    // width 4
	}
	line2 := []RenderedElement{
		{Primary: "XXXXXXXXX", BgColor: "#a00"}, // width 11
		{Primary: "Y", BgColor: "#0a0"},         // width 3
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	b0 := g1Boundary(ew, result.Pads, 0)
	b1 := g1Boundary(ew, result.Pads, 1)
	b2 := g1Boundary(ew, result.Pads, 2)
	if b0 != b1 || b1 != b2 {
		t.Errorf("3-line penalty: G1 should align all 3, got L0=%d L1=%d L2=%d", b0, b1, b2)
	}
	if b0 != 11 {
		t.Errorf("3-line penalty: G1 target should be 11, got %d", b0)
	}
}

// TestSmartAlign_Edge_SpreadExactlyMaxAlignSpread verifies that a boundary
// at exactly maxAlignSpread (16) from the window start IS included.
func TestSmartAlign_Edge_SpreadExactlyMaxAlignSpread(t *testing.T) {
	// G1 positions: 4, 10, 20 (spread 4→20 = 16 = maxAlignSpread).
	// All 3 in window. maxLinePad=16 → penalty → effectiveLines=2.
	// 2-line sub-cluster {4,10}: spread=6, effectiveLines=2.
	// Both candidates have effectiveLines=2; algorithm picks the one with
	// better cost (3-line includes L2, giving it alignment).
	line0 := []RenderedElement{
		{Primary: "XX", BgColor: "#a00"}, // width 4
		{Primary: "Y", BgColor: "#0a0"},  // width 3
	}
	line1 := []RenderedElement{
		{Primary: "XXXXXXXX", BgColor: "#a00"}, // width 10
		{Primary: "Y", BgColor: "#0a0"},        // width 3
	}
	line2 := []RenderedElement{
		{Primary: strings.Repeat("X", 18), BgColor: "#a00"}, // width 20
		{Primary: "Y", BgColor: "#0a0"},                     // width 3
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	b0 := g1Boundary(ew, result.Pads, 0)
	b1 := g1Boundary(ew, result.Pads, 1)
	// At minimum, L0 and L1 must align (spread=6 ≤ maxGroupPad).
	if b0 != b1 {
		t.Errorf("spread=16: L0 and L1 must align, got L0=%d L1=%d", b0, b1)
	}
}

// TestSmartAlign_Edge_SpreadExceedsMaxAlignSpread verifies that a boundary
// at maxAlignSpread+1 (17) is excluded from the window, leaving only a
// partial 2-line alignment.
func TestSmartAlign_Edge_SpreadExceedsMaxAlignSpread(t *testing.T) {
	// G1 positions: 4, 10, 21 (spread 4→21 = 17 > maxAlignSpread).
	// Window from 4: {4,10} only (21 excluded). Spread=6 → aligned.
	// L2 alone → no cluster.
	line0 := []RenderedElement{
		{Primary: "XX", BgColor: "#a00"}, // width 4
		{Primary: "Y", BgColor: "#0a0"},  // width 3
	}
	line1 := []RenderedElement{
		{Primary: "XXXXXXXX", BgColor: "#a00"}, // width 10
		{Primary: "Y", BgColor: "#0a0"},        // width 3
	}
	line2 := []RenderedElement{
		{Primary: strings.Repeat("X", 19), BgColor: "#a00"}, // width 21
		{Primary: "Y", BgColor: "#0a0"},                     // width 3
	}
	allLines := [][]RenderedElement{line0, line1, line2}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	b0 := g1Boundary(ew, result.Pads, 0)
	b1 := g1Boundary(ew, result.Pads, 1)
	b2 := g1Boundary(ew, result.Pads, 2)
	// L0 and L1 aligned to 10.
	if b0 != b1 {
		t.Errorf("spread=17: L0 and L1 should align, got L0=%d L1=%d", b0, b1)
	}
	if b0 != 10 {
		t.Errorf("spread=17: L0+L1 target should be 10, got %d", b0)
	}
	// L2 stays at 21 (no alignment partner).
	if b2 != 21 {
		t.Errorf("spread=17: L2 should stay at 21, got %d", b2)
	}
}

// TestSmartAlign_Edge_G2NaturallyAlignedAfterG1 verifies that when G1
// alignment padding exactly compensates for G2 position differences, no
// additional G2 padding is needed.
func TestSmartAlign_Edge_G2NaturallyAlignedAfterG1(t *testing.T) {
	// Line 0: G1=5, G2=14. Line 1: G1=7, G2=16.
	// G1 aligns to 7, L0 gets +2.
	// G2 adjusted: L0=14+2=16, L1=16. Already equal → zero G2 padding.
	line0 := []RenderedElement{
		{Primary: "XXX", BgColor: "#a00"},     // width 5
		{Primary: "XXXXXXX", BgColor: "#0a0"}, // width 9. G2 pos=5+9=14
	}
	line1 := []RenderedElement{
		{Primary: "XXXXX", BgColor: "#a00"},   // width 7
		{Primary: "XXXXXXX", BgColor: "#0a0"}, // width 9. G2 pos=7+9=16
	}
	allLines := [][]RenderedElement{line0, line1}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	// G1: L0 should get +2, L1 should get 0.
	if result.Pads[0][0] != 2 {
		t.Errorf("G1 pad L0: want 2, got %d", result.Pads[0][0])
	}
	if result.Pads[1][0] != 0 {
		t.Errorf("G1 pad L1: want 0, got %d", result.Pads[1][0])
	}
	// G2: no alignment padding (zero trailing fill too — both at 16).
	if result.Pads[0][1] != 0 {
		t.Errorf("G2 pad L0: want 0, got %d", result.Pads[0][1])
	}
	if result.Pads[1][1] != 0 {
		t.Errorf("G2 pad L1: want 0, got %d", result.Pads[1][1])
	}
}

// TestSmartAlign_Edge_ComplementaryWidths verifies alignment when G1 and G2
// widths are swapped between lines (total line widths equal).
func TestSmartAlign_Edge_ComplementaryWidths(t *testing.T) {
	// Line 0: G1=5, G2=9 → G2 pos=14, lineWidth=14
	// Line 1: G1=9, G2=5 → G2 pos=14, lineWidth=14
	// G1: spread=4, aligns to 9. L0+4.
	// G2: L0 adj=14+4=18, L1 adj=14. Target=18. L1+4.
	// Both levels get +4 on the shorter side. Total widths: 14+4=18 each.
	line0 := []RenderedElement{
		{Primary: "XXX", BgColor: "#a00"},     // width 5
		{Primary: "XXXXXXX", BgColor: "#0a0"}, // width 9
	}
	line1 := []RenderedElement{
		{Primary: "XXXXXXX", BgColor: "#a00"}, // width 9
		{Primary: "XXX", BgColor: "#0a0"},     // width 5
	}
	allLines := [][]RenderedElement{line0, line1}
	ew := testElemWidths(allLines)
	result := SmartAlign(allLines, ew, "", 0)
	smartAlignInvariants(t, allLines, ew, "", 0, result)

	// G1 boundaries aligned.
	b0 := g1Boundary(ew, result.Pads, 0)
	b1 := g1Boundary(ew, result.Pads, 1)
	if b0 != b1 || b0 != 9 {
		t.Errorf("G1 should align to 9, got L0=%d L1=%d", b0, b1)
	}

	// G2 boundaries aligned.
	g2b0 := ew[0][0] + result.Pads[0][0] + ew[0][1] + result.Pads[0][1]
	g2b1 := ew[1][0] + result.Pads[1][0] + ew[1][1] + result.Pads[1][1]
	if g2b0 != g2b1 || g2b0 != 18 {
		t.Errorf("G2 should align to 18, got L0=%d L1=%d", g2b0, g2b1)
	}
}
