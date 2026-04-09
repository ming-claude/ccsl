package render

import (
	"fmt"
	"strings"
)

// maxAlignSpread is the maximum position spread within an alignment cluster.
const maxAlignSpread = 16

// maxGroupPad is the maximum padding a single group can receive from one
// cluster before the cluster's effective coverage is demoted by 1 line.
// This prevents the algorithm from adding huge visual gaps that displace
// adjacent content (especially variable-width segments like cwd and git).
const maxGroupPad = 6

// AlignResult holds the output of SmartAlign: per-element padding and optional
// per-element shrinkage deltas.
type AlignResult struct {
	Pads    [][]int // per-element trailing padding (includes trailing fill)
	Shrinks [][]int // per-element shrinkage delta (columns to remove); nil if no shrinkage
}

// boundary records where a colored block ends on a particular line.
type boundary struct {
	lineIdx     int // which line
	blockIdx    int // which block on that line (used to find last element)
	lastElemIdx int // index in original elements slice of the last element in block
	position    int // cumulative display position after this block
}

// SmartAlign computes per-element trailing padding to align colored block
// boundaries across multiple lines. elemWidths[li][ei] carries the actual
// post-budget display width of each element (including block padding for
// colored elements). defaultSep is the separator string between blocks
// (used for accurate position calculation). maxWidth is the terminal width
// limit (0 = unlimited).
//
// Algorithm:
//  1. Build all-zero padding result.
//  2. Single line or empty → return zeros.
//  3. For each line, iterate through color blocks, accumulate position
//     (including separator widths between blocks).
//     Record boundary after each block that has a BgColor.
//  4. Sort boundaries by position ascending.
//  5. Greedy loop (merged clustering + apply): repeatedly pick the window
//     (position span ≤ maxAlignSpread) covering the most distinct lines;
//     break ties by lowest variance (P2.1) then lowest total padding (P2.2).
//     After picking, immediately apply padding and update running state.
func SmartAlign(allLines [][]RenderedElement, elemWidths [][]int, defaultSep string, maxWidth int) AlignResult {
	// Step 1: initialise result.
	pads := make([][]int, len(allLines))
	for i, line := range allLines {
		pads[i] = make([]int, len(line))
	}

	result := AlignResult{Pads: pads}

	// Step 2: trivial cases.
	if len(allLines) <= 1 {
		return result
	}

	// Detect non-block mode: when no element has BgColor, align separator
	// positions instead of colored block boundaries.
	nonBlockMode := true
	for _, line := range allLines {
		for _, e := range line {
			if e.BgColor != "" {
				nonBlockMode = false
				break
			}
		}
		if !nonBlockMode {
			break
		}
	}

	// Step 3: build boundaries, compute per-line widths and group counts.
	var boundaries []boundary
	lineWidths := make([]int, len(allLines))
	// groupCounts[li] = number of boundary-generating blocks on line li.
	groupCounts := make([]int, len(allLines))
	// lastBlockIdx[li] = blockIdx of last boundary block on line li (-1 if none).
	lastBlockIdx := make([]int, len(allLines))
	for i := range lastBlockIdx {
		lastBlockIdx[i] = -1
	}

	for li, line := range allLines {
		allVisible := make([]bool, len(line))
		for i := range allVisible {
			allVisible[i] = true
		}
		blocks := mergeSameColorBlocks(line, allVisible)

		pos := 0
		groupIdx := 0
		for bi, blk := range blocks {
			// Add separator width from previous block.
			if bi > 0 {
				prevBlk := blocks[bi-1]
				lastPrevIdx := prevBlk.indices[len(prevBlk.indices)-1]
				sep := defaultSep
				if line[lastPrevIdx].SeparatorAfter != "" {
					sep = line[lastPrevIdx].SeparatorAfter
				}
				pos += DisplayWidth(sep)
			}

			// Compute block width from elemWidths.
			blockWidth := 0
			for _, idx := range blk.indices {
				blockWidth += elemWidths[li][idx]
			}
			pos += blockWidth

			if blk.bgColor != "" || nonBlockMode {
				lastIdx := blk.indices[len(blk.indices)-1]
				boundaries = append(boundaries, boundary{
					lineIdx:     li,
					blockIdx:    groupIdx,
					lastElemIdx: lastIdx,
					position:    pos,
				})
				lastBlockIdx[li] = groupIdx
				groupIdx++
			}
		}
		lineWidths[li] = pos
		groupCounts[li] = groupIdx
	}

	totalGroups := 0
	for _, c := range groupCounts {
		totalGroups += c
	}

	if totalGroups == 0 || len(boundaries) == 0 {
		return result
	}

	// Step 4: sort by position ascending.
	sortBoundaries(boundaries)

	// Step 5: greedy loop with variance-aware cost (width-aware with shrinkage).
	padOffset, shrinks := greedyAlignWithVariance(
		boundaries, lineWidths, groupCounts, lastBlockIdx, totalGroups,
		pads, allLines, elemWidths, maxWidth, nonBlockMode,
	)
	result.Shrinks = shrinks

	// Compute per-line shrink offset from shrinks.
	shrinkOff := make([]int, len(allLines))
	if result.Shrinks != nil {
		for li, ls := range result.Shrinks {
			for _, s := range ls {
				shrinkOff[li] += s
			}
		}
	}

	// Step 6: trailing fill — equalize all lines to the same total width.
	// Skip for non-block mode: trailing spaces are invisible without BgColor.
	if !nonBlockMode {
		// Build lastElemOfLastBlock lookup from boundaries.
		lastElemOfLastBlock := make([]int, len(allLines))
		for i := range lastElemOfLastBlock {
			lastElemOfLastBlock[i] = -1
		}
		for _, b := range boundaries {
			if b.blockIdx == lastBlockIdx[b.lineIdx] {
				lastElemOfLastBlock[b.lineIdx] = b.lastElemIdx
			}
		}

		// Find the widest line (after alignment + shrinkage).
		maxTotalWidth := 0
		for li := range lineWidths {
			if groupCounts[li] == 0 {
				continue
			}
			w := lineWidths[li] + padOffset[li] - shrinkOff[li]
			if w > maxTotalWidth {
				maxTotalWidth = w
			}
		}

		// Cap at maxWidth (R4 takes priority over R2).
		if maxWidth > 0 && maxTotalWidth > maxWidth {
			maxTotalWidth = maxWidth
		}

		// Apply trailing fill: pad each line's last colored block to reach maxTotalWidth.
		for li := range lineWidths {
			if groupCounts[li] == 0 || lastElemOfLastBlock[li] < 0 {
				continue
			}
			w := lineWidths[li] + padOffset[li] - shrinkOff[li]
			fill := maxTotalWidth - w
			if fill > 0 {
				pads[li][lastElemOfLastBlock[li]] += fill
			}
		}
	}

	return result
}

// sortBoundaries sorts boundaries by position ascending using insertion sort.
func sortBoundaries(bs []boundary) {
	for i := 1; i < len(bs); i++ {
		key := bs[i]
		j := i - 1
		for j >= 0 && bs[j].position > key.position {
			bs[j+1] = bs[j]
			j--
		}
		bs[j+1] = key
	}
}

// greedyAlignWithVariance is the merged greedy clustering + apply step.
// It selects clusters one at a time, immediately applying padding after each
// selection, so that subsequent evaluations see the updated state.
//
// Cost function priority (P1 > P2.1 > P2.2 > P2.3):
//   - P1: nLines (more is better)
//   - P2.1: sumSqDev — variance proxy across all groups (lower is better)
//   - P2.2: totalPad including trailing fill (lower is better)
//   - P2.3: totalShrink (lower is better)
//
// When maxWidth > 0, clusters that would push any line beyond maxWidth are
// checked for shrinkage feasibility. Variable-width elements (IsVariable=true)
// can be shrunk down to MinWidth to absorb overflow.
func greedyAlignWithVariance(
	boundaries []boundary,
	lineWidths []int,
	groupCounts []int,
	lastBlockIdx []int,
	totalGroups int,
	pads [][]int,
	allLines [][]RenderedElement,
	elemWidths [][]int,
	maxWidth int,
	nonBlockMode bool,
) ([]int, [][]int) {
	numLines := len(lineWidths)
	assigned := make([]bool, len(boundaries))
	padOffset := make([]int, numLines)
	shrinkOffset := make([]int, numLines)

	// Per-element shrink capacity: elemWidth - MinWidth for variable elements.
	shrinkRemaining := make([][]int, numLines)
	shrinks := make([][]int, numLines)
	for li, line := range allLines {
		shrinkRemaining[li] = make([]int, len(line))
		shrinks[li] = make([]int, len(line))
		for ei, e := range line {
			if e.IsVariable && e.MinWidth > 0 && elemWidths[li][ei] > e.MinWidth {
				shrinkRemaining[li][ei] = elemWidths[li][ei] - e.MinWidth
			}
		}
	}

	// groupPads[li][blockIdx] = current padding for that colored group.
	groupPads := make([][]int, numLines)
	for li := range groupPads {
		groupPads[li] = make([]int, groupCounts[li])
	}

	// shrinkUpTo returns the total shrinkage on elements at or before upToIdx
	// on line li. Shrinkage AFTER a boundary does not shift that boundary's
	// visual position, so boundary position calculations must use this instead
	// of the per-line shrinkOffset total.
	shrinkUpTo := func(li, upToIdx int) int {
		total := 0
		for ei := 0; ei <= upToIdx && ei < len(shrinks[li]); ei++ {
			total += shrinks[li][ei]
		}
		return total
	}

	// computeStats computes Σ(pad[g] - mean)² (ssd) and total pad across all
	// groups, including trailing fill attributed to each line's last colored
	// group (skipped in non-block mode). maxWidthCap caps the effective max
	// line width (0 = no cap).
	computeStats := func(gp [][]int, po, so []int, maxWidthCap int) (ssd, totalPad int) {
		maxW := 0
		if !nonBlockMode {
			for li := range lineWidths {
				if groupCounts[li] == 0 {
					continue
				}
				w := lineWidths[li] + po[li] - so[li]
				if w > maxW {
					maxW = w
				}
			}
			if maxWidthCap > 0 && maxW > maxWidthCap {
				maxW = maxWidthCap
			}
		}

		totalPad = 0
		for li := range gp {
			for bi, p := range gp[li] {
				v := p
				if !nonBlockMode && bi == lastBlockIdx[li] {
					fill := maxW - (lineWidths[li] + po[li] - so[li])
					if fill > 0 {
						v += fill
					}
				}
				totalPad += v
			}
		}
		if totalGroups == 0 {
			return 0, totalPad
		}

		// Use scaled deviation (v*N - totalPad) instead of (v - totalPad/N)
		// to avoid integer division truncation. The N² scaling factor is
		// constant across all candidates, so relative ordering is preserved.
		ssd = 0
		for li := range gp {
			for bi, p := range gp[li] {
				v := p
				if !nonBlockMode && bi == lastBlockIdx[li] {
					fill := maxW - (lineWidths[li] + po[li] - so[li])
					if fill > 0 {
						v += fill
					}
				}
				d := v*totalGroups - totalPad
				ssd += d * d
			}
		}
		return
	}

	// totalShrinkCapacity returns the sum of all remaining shrink capacity.
	totalShrinkCapacity := func() int {
		total := 0
		for li := range shrinkRemaining {
			for _, r := range shrinkRemaining[li] {
				total += r
			}
		}
		return total
	}

	// Determine the maximum blockIdx across all boundaries so we can
	// iterate level by level (G1 → G2 → G3 …). Processing earlier block
	// levels first ensures that padding from G1 alignment is settled before
	// G2 boundaries are evaluated — otherwise G1 padding applied later would
	// shift already-aligned G2 boundaries.
	maxBI := 0
	for _, b := range boundaries {
		if b.blockIdx > maxBI {
			maxBI = b.blockIdx
		}
	}

	for currentBI := 0; currentBI <= maxBI; currentBI++ {
		for {
			var free []int
			for i, a := range assigned {
				if !a && boundaries[i].blockIdx == currentBI {
					free = append(free, i)
				}
			}
			if len(free) == 0 {
				break
			}

			bestLines := 0
			bestSSD := int(^uint(0) >> 1) // max int
			bestPad := int(^uint(0) >> 1)
			bestShrinkAmt := int(^uint(0) >> 1)
			var bestWindow []int

			for si := 0; si < len(free); si++ {
				// Window/spread evaluation uses original boundary positions (stable).
				startPos := boundaries[free[si]].position

				lastPosPerLine := map[int]int{}
				var window []int

				for ei := si; ei < len(free); ei++ {
					idx := free[ei]
					pos := boundaries[idx].position
					if pos-startPos > maxAlignSpread {
						break
					}
					li := boundaries[idx].lineIdx
					if _, ok := lastPosPerLine[li]; ok {
						continue // one boundary per line per window — prevents cross-block contamination
					}
					window = append(window, idx)
					lastPosPerLine[li] = pos
				}

				nLines := len(lastPosPerLine)
				if nLines < 2 {
					continue
				}

				// Simulate apply on temporary copies.
				tmpOffset := make([]int, numLines)
				copy(tmpOffset, padOffset)
				tmpShrinkOffset := make([]int, numLines)
				copy(tmpShrinkOffset, shrinkOffset)
				tmpPads := make([][]int, numLines)
				for li := range groupPads {
					tmpPads[li] = make([]int, len(groupPads[li]))
					copy(tmpPads[li], groupPads[li])
				}

				// Target computation uses adjusted positions.
				last := map[int]boundary{}
				for _, idx := range window {
					b := boundaries[idx]
					if existing, ok := last[b.lineIdx]; !ok || b.position > existing.position {
						last[b.lineIdx] = b
					}
				}

				clusterTarget := 0
				for li, b := range last {
					adj := b.position + tmpOffset[li] - shrinkUpTo(li, b.lastElemIdx)
					if adj > clusterTarget {
						clusterTarget = adj
					}
				}
				maxLinePad := 0
				for li, b := range last {
					adj := b.position + tmpOffset[li] - shrinkUpTo(li, b.lastElemIdx)
					need := clusterTarget - adj
					if need > maxLinePad {
						maxLinePad = need
					}
					if need > 0 {
						tmpOffset[li] += need
						tmpPads[li][b.blockIdx] += need
					}
				}

				// Displacement penalty: if the largest per-line padding exceeds
				// maxGroupPad, demote effective coverage by 1. This prevents the
				// algorithm from creating huge visual gaps that squeeze adjacent
				// content (especially variable-width segments).
				effectiveLines := nLines
				if maxLinePad > maxGroupPad {
					effectiveLines--
				}
				if effectiveLines < 2 {
					continue // too much displacement, skip
				}

				// Check width constraint and compute shrinkage needed.
				candidateShrink := 0
				widthCap := 0
				if maxWidth > 0 {
					maxEff := 0
					for li := range lineWidths {
						if groupCounts[li] == 0 {
							continue
						}
						w := lineWidths[li] + tmpOffset[li] - tmpShrinkOffset[li]
						if w > maxEff {
							maxEff = w
						}
					}
					if maxEff > maxWidth {
						overflow := maxEff - maxWidth
						if totalShrinkCapacity() < overflow {
							continue // infeasible, skip
						}
						candidateShrink = overflow
						widthCap = maxWidth
					}
				}

				ssd, totalPadIncl := computeStats(tmpPads, tmpOffset, tmpShrinkOffset, widthCap)

				// Four-level comparison: P1 effectiveLines↑, P2.1 ssd↓, P2.2 totalPad↓, P2.3 shrink↓.
				if effectiveLines > bestLines ||
					(effectiveLines == bestLines && ssd < bestSSD) ||
					(effectiveLines == bestLines && ssd == bestSSD && totalPadIncl < bestPad) ||
					(effectiveLines == bestLines && ssd == bestSSD && totalPadIncl == bestPad && candidateShrink < bestShrinkAmt) {
					bestLines = effectiveLines
					bestSSD = ssd
					bestPad = totalPadIncl
					bestShrinkAmt = candidateShrink
					bestWindow = append([]int(nil), window...)
				}
			}

			if bestWindow == nil {
				break
			}

			// Actually apply the winning cluster's padding.
			last := map[int]boundary{}
			for _, idx := range bestWindow {
				assigned[idx] = true
				b := boundaries[idx]
				if existing, ok := last[b.lineIdx]; !ok || b.position > existing.position {
					last[b.lineIdx] = b
				}
			}

			clusterTarget := 0
			for li, b := range last {
				adj := b.position + padOffset[li] - shrinkUpTo(li, b.lastElemIdx)
				if adj > clusterTarget {
					clusterTarget = adj
				}
			}
			for li, b := range last {
				adj := b.position + padOffset[li] - shrinkUpTo(li, b.lastElemIdx)
				need := clusterTarget - adj
				if need > 0 {
					pads[li][b.lastElemIdx] += need
					padOffset[li] += need
					groupPads[li][b.blockIdx] += need
				}
			}

			// Apply shrinkage if this cluster causes overflow.
			if maxWidth > 0 {
				maxEff := 0
				for li := range lineWidths {
					if groupCounts[li] == 0 {
						continue
					}
					w := lineWidths[li] + padOffset[li] - shrinkOffset[li]
					if w > maxEff {
						maxEff = w
					}
				}
				if maxEff > maxWidth {
					distributeShrinkage(allLines, lineWidths, padOffset, shrinkOffset,
						shrinkRemaining, shrinks, groupCounts, maxWidth)
				}
			}
		}
	} // end for currentBI

	// Return nil shrinks if no shrinkage was applied.
	anyShrink := false
	for _, ls := range shrinks {
		for _, s := range ls {
			if s > 0 {
				anyShrink = true
				break
			}
		}
		if anyShrink {
			break
		}
	}
	if !anyShrink {
		shrinks = nil
	}

	return padOffset, shrinks
}

// distributeShrinkage shrinks variable-width elements on every line that
// exceeds maxWidth, each by exactly the amount needed. This correctly handles
// the case where multiple lines are at the same effective width — each line
// computes its own overflow independently.
func distributeShrinkage(
	allLines [][]RenderedElement,
	lineWidths, padOffset, shrinkOffset []int,
	shrinkRemaining, shrinks [][]int,
	groupCounts []int,
	maxWidth int,
) {
	for li := range allLines {
		if groupCounts[li] == 0 {
			continue
		}
		w := lineWidths[li] + padOffset[li] - shrinkOffset[li]
		if w <= maxWidth {
			continue
		}
		overflow := w - maxWidth
		for ei := range allLines[li] {
			if overflow <= 0 {
				break
			}
			r := shrinkRemaining[li][ei]
			if r <= 0 {
				continue
			}
			take := r
			if take > overflow {
				take = overflow
			}
			shrinkOffset[li] += take
			shrinkRemaining[li][ei] -= take
			shrinks[li][ei] += take
			overflow -= take
		}
	}
}

// colorBlock groups consecutive visible element indices that share the same
// background color. Elements without a BgColor each form their own block with
// bgColor == "".
type colorBlock struct {
	indices []int
	bgColor string
}

// mergeSameColorBlocks groups consecutive visible elements that share the same
// non-empty BgColor into a single colorBlock. Elements without BgColor are
// each placed in their own block. Hidden elements (visible[i] == false) break
// adjacency — they are not included in any block, but their presence prevents
// the surrounding visible elements from being merged.
func mergeSameColorBlocks(elements []RenderedElement, visible []bool) []colorBlock {
	var blocks []colorBlock

	for i, e := range elements {
		if !visible[i] {
			// Hidden element breaks adjacency: any open same-color run ends here.
			// We signal this by setting the last block's bgColor sentinel so the
			// next visible element cannot merge into it.  The simplest approach is
			// to nil-out the current run by appending a sentinel-free tail — but
			// since we only check the last block's bgColor, we just need to ensure
			// the next element can't merge.  We use a flag to track the break.
			if len(blocks) > 0 {
				// Mark the last block as "closed" by clearing its bgColor so the
				// next visible element with the same color starts a new block.
				// We achieve this by recording that a hidden element was encountered
				// after the last block — we do this by appending a zero-length
				// sentinel block with an impossible bgColor that no real element
				// can match.
				blocks = append(blocks, colorBlock{bgColor: "\x00"})
			}
			continue
		}

		bg := e.BgColor
		if bg != "" && len(blocks) > 0 {
			last := &blocks[len(blocks)-1]
			if last.bgColor == bg {
				// Same color as the previous block — extend it.
				last.indices = append(last.indices, i)
				continue
			}
		}

		// Start a new block.
		blocks = append(blocks, colorBlock{
			indices: []int{i},
			bgColor: bg,
		})
	}

	// Remove the zero-length sentinel blocks (those with no indices).
	out := blocks[:0]
	for _, b := range blocks {
		if len(b.indices) > 0 {
			out = append(out, b)
		}
	}
	return out
}

// assembleWithMerge assembles the line like assembleVisible, but merges
// consecutive visible elements that share the same non-empty BgColor into a
// single ColorizeBlock.
//
// pads is an optional per-element padding width (extra spaces appended to the
// raw text before re-colorizing).  Pass nil for no padding.
func assembleWithMerge(elements []RenderedElement, texts []string, defaultSep string, visible []bool, pads []int) string {
	blocks := mergeSameColorBlocks(elements, visible)
	if len(blocks) == 0 {
		return ""
	}

	var b strings.Builder

	for bi, blk := range blocks {
		// Separator before this block (from the last element of the previous block).
		if bi > 0 {
			prevBlk := blocks[bi-1]
			lastPrevIdx := prevBlk.indices[len(prevBlk.indices)-1]
			sep := defaultSep
			if elements[lastPrevIdx].SeparatorAfter != "" {
				sep = elements[lastPrevIdx].SeparatorAfter
			}
			b.WriteString(colorizeSep(sep, elements[lastPrevIdx]))
		}

		if blk.bgColor != "" && len(blk.indices) > 1 {
			// Multi-element same-color block: strip individual ANSI, join with
			// double space, add optional padding to the last element, then
			// re-colorize as one block.
			var parts []string
			for _, idx := range blk.indices {
				raw := strings.TrimSpace(StripANSI(texts[idx]))
				parts = append(parts, raw)
			}
			joined := strings.Join(parts, "  ")

			// Append alignment padding to the last element if requested.
			if pads != nil {
				lastIdx := blk.indices[len(blk.indices)-1]
				if lastIdx < len(pads) && pads[lastIdx] > 0 {
					joined += strings.Repeat(" ", pads[lastIdx])
				}
			}

			// Use the bg and fg of the first element in the block.
			fg := elements[blk.indices[0]].Color
			b.WriteString(ColorizeBlock(joined, fg, blk.bgColor))
		} else if blk.bgColor != "" {
			// Single-element block with bg color.
			idx := blk.indices[0]
			raw := strings.TrimSpace(StripANSI(texts[idx]))

			if pads != nil && idx < len(pads) && pads[idx] > 0 {
				raw += strings.Repeat(" ", pads[idx])
			}

			fg := elements[idx].Color
			b.WriteString(ColorizeBlock(raw, fg, blk.bgColor))
		} else {
			// Single-element block without bg color: use text as-is.
			idx := blk.indices[0]
			b.WriteString(texts[idx])
		}
	}

	return b.String()
}

// AlignDiagram holds a simplified text visualization of SmartAlign's effect.
// Each line is rendered as alternating block characters (█/░) for colored blocks,
// with · representing padding injected by SmartAlign.
type AlignDiagram struct {
	Before []string `json:"before"` // pre-alignment layout
	After  []string `json:"after"`  // post-alignment layout with padding shown as ·
}

// BuildAlignDiagram generates a before/after visualization of SmartAlign's
// effect on block layout. It uses the same inputs and result as SmartAlign.
// Returns nil if there are fewer than 2 lines or no colored blocks.
func BuildAlignDiagram(
	allLines [][]RenderedElement,
	elemWidths [][]int,
	result AlignResult,
	defaultSep string,
) *AlignDiagram {
	if len(allLines) <= 1 {
		return nil
	}

	// Check if any colored blocks exist.
	hasColored := false
	for _, line := range allLines {
		for _, e := range line {
			if e.BgColor != "" {
				hasColored = true
				break
			}
		}
		if hasColored {
			break
		}
	}
	if !hasColored {
		return nil
	}

	before := make([]string, len(allLines))
	after := make([]string, len(allLines))

	for li, line := range allLines {
		allVisible := make([]bool, len(line))
		for i := range allVisible {
			allVisible[i] = true
		}
		blocks := mergeSameColorBlocks(line, allVisible)

		var beforeBuf, afterBuf strings.Builder
		colorIdx := 0

		for bi, blk := range blocks {
			// Block content width = sum of element widths.
			blockWidth := 0
			for _, idx := range blk.indices {
				blockWidth += elemWidths[li][idx]
			}

			// Separator width after this block (absorbed into this block's
			// visual representation if this block has a bgColor, since
			// colorizeSep colors it with the left block's bg).
			sepWidth := 0
			if bi < len(blocks)-1 {
				lastIdx := blk.indices[len(blk.indices)-1]
				sep := defaultSep
				if line[lastIdx].SeparatorAfter != "" {
					sep = line[lastIdx].SeparatorAfter
				}
				sepWidth = DisplayWidth(sep)
			}

			// Sum padding for this block's elements.
			blockPad := 0
			if li < len(result.Pads) {
				for _, idx := range blk.indices {
					if idx < len(result.Pads[li]) {
						blockPad += result.Pads[li][idx]
					}
				}
			}

			// Sum shrinkage for this block's elements.
			blockShrink := 0
			if result.Shrinks != nil && li < len(result.Shrinks) {
				for _, idx := range blk.indices {
					if idx < len(result.Shrinks[li]) {
						blockShrink += result.Shrinks[li][idx]
					}
				}
			}

			if blk.bgColor != "" {
				fillChar := '█'
				if colorIdx%2 == 1 {
					fillChar = '░'
				}

				// Before: block content + absorbed separator.
				totalBefore := blockWidth + sepWidth
				writeRunes(&beforeBuf, fillChar, totalBefore)

				// After: (content - shrink) + padding + absorbed separator.
				contentAfter := blockWidth - blockShrink
				writeRunes(&afterBuf, fillChar, contentAfter)
				writeRunes(&afterBuf, '·', blockPad)
				writeRunes(&afterBuf, fillChar, sepWidth)

				colorIdx++
			} else {
				// Non-colored element: show as space (with separator).
				totalWidth := blockWidth + sepWidth
				writeRunes(&beforeBuf, ' ', totalWidth)
				writeRunes(&afterBuf, ' ', totalWidth)
			}
		}

		before[li] = fmt.Sprintf("L%d: %s", li+1, beforeBuf.String())
		after[li] = fmt.Sprintf("L%d: %s", li+1, afterBuf.String())
	}

	return &AlignDiagram{Before: before, After: after}
}

// writeRunes writes r repeated n times to b. Negative n is a no-op.
func writeRunes(b *strings.Builder, r rune, n int) {
	for i := 0; i < n; i++ {
		b.WriteRune(r)
	}
}

// PadLineToWidth appends background-colored spaces to line until its display
// width reaches targetWidth. Returns line unchanged if:
//   - fillColor is empty
//   - targetWidth <= 0
//   - line is already at or above targetWidth
//   - fillColor is not a recognised color
func PadLineToWidth(line string, targetWidth int, fillColor string) string {
	if fillColor == "" || targetWidth <= 0 {
		return line
	}
	currentWidth := DisplayWidth(StripANSI(line))
	pad := targetWidth - currentWidth
	if pad <= 0 {
		return line
	}
	bgCode := colorCode(fillColor, true)
	if bgCode == "" {
		return line
	}
	return line + bgCode + strings.Repeat(" ", pad) + reset
}
