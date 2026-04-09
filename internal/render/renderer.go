package render

import "strings"

// RenderedElement is a single segment's rendered output ready for line assembly.
type RenderedElement struct {
	Icon           string
	Title          string
	Primary        string
	Detail         string
	Color          string // foreground color
	BgColor        string // background color (empty = text-only, no block)
	SeparatorAfter string // overrides default separator; empty = use default
	Priority       int    // 1-10, higher = kept longer when over budget
	MinWidth       int    // minimum width before hiding (variable-width only)
	IsVariable     bool   // can be truncated
}

// buildElementText builds the plain text for an element: "{icon} {title}{primary}{ detail}".
// Empty parts are skipped. Spacing rules:
//   - icon gets a trailing space if anything follows
//   - title and primary are concatenated directly
//   - detail gets a leading space
func buildElementText(e RenderedElement) string {
	var b strings.Builder

	if e.Icon != "" {
		b.WriteString(e.Icon)
		if e.Title != "" || e.Primary != "" || e.Detail != "" {
			b.WriteByte(' ')
		}
	}

	b.WriteString(e.Title)
	b.WriteString(e.Primary)

	if e.Detail != "" {
		b.WriteByte(' ')
		b.WriteString(e.Detail)
	}

	return b.String()
}

// budgetResult holds the per-line state produced by budgetLine.
type budgetResult struct {
	texts    []string
	visible  []bool
	hasBlock bool
}

// budgetLine builds colorized text for each element and applies the width-
// budgeting algorithm: truncate variable-width elements, then hide by ascending
// priority. The result contains the (possibly truncated) texts and a visibility
// mask. Both RenderLine and RenderAllLines delegate to this function.
func budgetLine(elements []RenderedElement, defaultSep string, maxWidth int) budgetResult {
	texts := make([]string, len(elements))
	hasBlock := false
	for i, e := range elements {
		raw := buildElementText(e)
		if e.BgColor != "" {
			texts[i] = ColorizeBlock(raw, e.Color, e.BgColor)
			hasBlock = true
		} else {
			texts[i] = Colorize(raw, e.Color)
		}
	}

	visible := make([]bool, len(elements))
	for i := range visible {
		visible[i] = true
	}

	if maxWidth <= 0 {
		return budgetResult{texts: texts, visible: visible, hasBlock: hasBlock}
	}

	defaultSepWidth := DisplayWidth(defaultSep)
	totalWidth := calcTotalWidth(elements, texts, defaultSep, defaultSepWidth)
	if totalWidth <= maxWidth {
		return budgetResult{texts: texts, visible: visible, hasBlock: hasBlock}
	}

	// First pass: truncate variable-width elements.
	for i, e := range elements {
		if !e.IsVariable {
			continue
		}
		textWidth := DisplayWidth(texts[i])
		if textWidth <= e.MinWidth {
			continue
		}
		reduction := textWidth - e.MinWidth
		if totalWidth-reduction <= maxWidth {
			needed := totalWidth - maxWidth
			targetWidth := textWidth - needed
			raw := buildElementText(e)
			truncated := Truncate(StripANSI(raw), targetWidth)
			if e.BgColor != "" {
				texts[i] = ColorizeBlock(truncated, e.Color, e.BgColor)
			} else {
				texts[i] = Colorize(truncated, e.Color)
			}
			totalWidth -= needed
			break
		}
		raw := buildElementText(e)
		truncated := Truncate(StripANSI(raw), e.MinWidth)
		if e.BgColor != "" {
			texts[i] = ColorizeBlock(truncated, e.Color, e.BgColor)
		} else {
			texts[i] = Colorize(truncated, e.Color)
		}
		totalWidth -= reduction
	}

	if totalWidth <= maxWidth {
		return budgetResult{texts: texts, visible: visible, hasBlock: hasBlock}
	}

	// Second pass: hide elements by ascending priority (lowest dropped first).
	for totalWidth > maxWidth {
		lowestIdx := -1
		lowestPri := int(^uint(0) >> 1)
		for i, e := range elements {
			if !visible[i] {
				continue
			}
			if e.Priority < lowestPri {
				lowestPri = e.Priority
				lowestIdx = i
			}
		}
		if lowestIdx < 0 {
			break
		}
		visible[lowestIdx] = false
		totalWidth = calcVisibleWidth(elements, texts, defaultSep, defaultSepWidth, visible)
	}

	return budgetResult{texts: texts, visible: visible, hasBlock: hasBlock}
}

// assembleBudgeted assembles a line from budgeted state, with optional pads
// for cross-line alignment. If the result still exceeds maxWidth, it is
// truncated as a last resort.
func assembleBudgeted(elements []RenderedElement, b budgetResult, defaultSep string, maxWidth int, pads []int) string {
	var line string
	if b.hasBlock {
		line = assembleWithMerge(elements, b.texts, defaultSep, b.visible, pads)
	} else {
		line = assembleVisible(elements, b.texts, defaultSep, b.visible, pads)
	}
	if maxWidth > 0 && DisplayWidth(line) > maxWidth {
		line = Truncate(line, maxWidth)
	}
	return line
}

// RenderLine assembles rendered elements into a single line respecting a width
// budget. When maxWidth <= 0 or the full line fits, all elements are included.
// When over budget the algorithm:
//  1. Truncates variable-width elements down to their MinWidth.
//  2. Hides elements by ascending priority (lowest first).
//  3. As a last resort, truncates the entire assembled line.
func RenderLine(elements []RenderedElement, defaultSep string, maxWidth int) string {
	if len(elements) == 0 {
		return ""
	}
	b := budgetLine(elements, defaultSep, maxWidth)
	return assembleBudgeted(elements, b, defaultSep, maxWidth, nil)
}

// RenderAllLines renders multiple lines with cross-line smart alignment and
// optional uniform-width fill. It uses budgetLine for per-line width budgeting,
// then SmartAlign to compute cross-line padding before final assembly.
//
// Parameters:
//   - lineElements: one slice of RenderedElement per line
//   - defaultSep:   separator string between elements (same for all lines)
//   - maxWidth:     per-line width budget (0 = unlimited)
func RenderAllLines(lineElements [][]RenderedElement, defaultSep string, maxWidth int) ([]string, *AlignDiagram) {
	if len(lineElements) == 0 {
		return nil, nil
	}

	// Per-line budget.
	budgets := make([]budgetResult, len(lineElements))
	for li, elements := range lineElements {
		if len(elements) == 0 {
			budgets[li] = budgetResult{}
			continue
		}
		budgets[li] = budgetLine(elements, defaultSep, maxWidth)
	}

	// Build the visible-elements view and compute post-budget display widths
	// for SmartAlign.
	alignInputs := make([][]RenderedElement, len(budgets))
	alignIndexMap := make([][]int, len(budgets))
	elemWidths := make([][]int, len(budgets))
	for li, b := range budgets {
		var visElems []RenderedElement
		var origIdxs []int
		var widths []int
		for i, e := range lineElements[li] {
			if b.visible == nil || b.visible[i] {
				visElems = append(visElems, e)
				origIdxs = append(origIdxs, i)
				widths = append(widths, DisplayWidth(b.texts[i]))
			}
		}
		alignInputs[li] = visElems
		alignIndexMap[li] = origIdxs
		elemWidths[li] = widths
	}

	alignResult := SmartAlign(alignInputs, elemWidths, defaultSep, maxWidth)

	// Build alignment diagram before padding absorption changes the pads.
	alignDiagram := BuildAlignDiagram(alignInputs, elemWidths, alignResult, defaultSep)

	// Apply shrink instructions: re-truncate affected elements' texts.
	if alignResult.Shrinks != nil {
		for li, lineShrinks := range alignResult.Shrinks {
			for visIdx, shrinkDelta := range lineShrinks {
				if shrinkDelta <= 0 {
					continue
				}
				origIdx := alignIndexMap[li][visIdx]
				e := lineElements[li][origIdx]
				newWidth := elemWidths[li][visIdx] - shrinkDelta
				if newWidth < 0 {
					newWidth = 0
				}
				// Strip ANSI and block padding before re-truncating.
				raw := strings.TrimSpace(StripANSI(budgets[li].texts[origIdx]))
				if e.BgColor != "" {
					// Reserve 2 columns for block padding added by ColorizeBlock.
					targetRaw := newWidth - 2
					if targetRaw < 0 {
						targetRaw = 0
					}
					truncated := Truncate(raw, targetRaw)
					budgets[li].texts[origIdx] = ColorizeBlock(truncated, e.Color, e.BgColor)
				} else {
					truncated := Truncate(raw, newWidth)
					budgets[li].texts[origIdx] = Colorize(truncated, e.Color)
				}
			}
		}
	}

	// Translate SmartAlign pads back to original element indices.
	padsPerLine := make([][]int, len(budgets))
	for li := range budgets {
		if len(lineElements[li]) == 0 {
			continue
		}
		p := make([]int, len(lineElements[li]))
		for newIdx, origIdx := range alignIndexMap[li] {
			if newIdx < len(alignResult.Pads[li]) {
				p[origIdx] = alignResult.Pads[li][newIdx]
			}
		}
		// Only pass pads if any are non-zero.
		anyPad := false
		for _, v := range p {
			if v != 0 {
				anyPad = true
				break
			}
		}
		if anyPad {
			padsPerLine[li] = p
		}
	}

	// Absorb alignment padding into truncated variable-width elements.
	// If a variable-width element was truncated by budgetLine and then
	// SmartAlign padded it, re-render the element wider using the padding
	// space. This replaces visual gaps with content while preserving total
	// block width and cross-line alignment.
	for li := range padsPerLine {
		if padsPerLine[li] == nil {
			continue
		}
		for i, pad := range padsPerLine[li] {
			if pad <= 0 {
				continue
			}
			e := lineElements[li][i]
			if !e.IsVariable {
				continue
			}
			currentWidth := DisplayWidth(budgets[li].texts[i])
			originalRaw := buildElementText(e)
			originalWidth := DisplayWidth(originalRaw)
			if e.BgColor != "" {
				originalWidth += 2 // block padding
			}
			if currentWidth >= originalWidth {
				continue // not truncated, nothing to absorb
			}
			// Absorb padding: widen content up to original width.
			absorb := pad
			if currentWidth+absorb > originalWidth {
				absorb = originalWidth - currentWidth
			}
			if absorb <= 0 {
				continue
			}
			// Re-render element at wider width from original text.
			if e.BgColor != "" {
				targetRaw := currentWidth + absorb - 2
				truncated := Truncate(originalRaw, targetRaw)
				budgets[li].texts[i] = ColorizeBlock(truncated, e.Color, e.BgColor)
			} else {
				truncated := Truncate(originalRaw, currentWidth+absorb)
				budgets[li].texts[i] = Colorize(truncated, e.Color)
			}
			padsPerLine[li][i] -= absorb
		}
	}

	// Assemble each line.
	result := make([]string, len(budgets))
	for li, b := range budgets {
		if len(lineElements[li]) == 0 {
			continue
		}
		result[li] = assembleBudgeted(lineElements[li], b, defaultSep, maxWidth, padsPerLine[li])
	}

	return result, alignDiagram
}

// calcTotalWidth computes the display width of all elements plus separators.
// defaultSepWidth is the pre-computed DisplayWidth of defaultSep to avoid
// recomputing it for every separator.
func calcTotalWidth(elements []RenderedElement, texts []string, defaultSep string, defaultSepWidth int) int {
	total := 0
	for i, t := range texts {
		total += DisplayWidth(t)
		if i < len(texts)-1 {
			if elements[i].SeparatorAfter != "" {
				total += DisplayWidth(elements[i].SeparatorAfter)
			} else {
				total += defaultSepWidth
			}
		}
	}
	return total
}

// calcVisibleWidth computes the display width of only visible elements.
// defaultSepWidth is the pre-computed DisplayWidth of defaultSep to avoid
// recomputing it for every separator.
func calcVisibleWidth(elements []RenderedElement, texts []string, defaultSep string, defaultSepWidth int, visible []bool) int {
	total := 0
	first := true
	prevIdx := -1
	for i := range texts {
		if !visible[i] {
			continue
		}
		if !first {
			if elements[prevIdx].SeparatorAfter != "" {
				total += DisplayWidth(elements[prevIdx].SeparatorAfter)
			} else {
				total += defaultSepWidth
			}
		}
		total += DisplayWidth(texts[i])
		first = false
		prevIdx = i
	}
	return total
}

// colorizeSep wraps a separator with the left element's bg color if present.
func colorizeSep(sep string, leftElem RenderedElement) string {
	if sep == "" || leftElem.BgColor == "" {
		return sep
	}
	bgCode := colorCode(leftElem.BgColor, true)
	if bgCode == "" {
		return sep
	}
	return bgCode + sep + reset
}

// assembleVisible joins only visible elements with separators.
// pads is an optional per-element padding width (extra spaces appended after
// element text for cross-line alignment). Pass nil for no padding.
func assembleVisible(elements []RenderedElement, texts []string, defaultSep string, visible []bool, pads []int) string {
	var b strings.Builder
	first := true
	prevIdx := -1
	for i, t := range texts {
		if !visible[i] {
			continue
		}
		if !first {
			sep := defaultSep
			if elements[prevIdx].SeparatorAfter != "" {
				sep = elements[prevIdx].SeparatorAfter
			}
			b.WriteString(colorizeSep(sep, elements[prevIdx]))
		}
		b.WriteString(t)
		if pads != nil && i < len(pads) && pads[i] > 0 {
			b.WriteString(strings.Repeat(" ", pads[i]))
		}
		first = false
		prevIdx = i
	}
	return b.String()
}
