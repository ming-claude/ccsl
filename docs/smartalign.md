# SmartAlign — 多行色块边界智能对齐算法

## 概述

SmartAlign 是 CCSL 的核心渲染算法，为多行 block 主题的 statusline 做跨行色块边界对齐。它在终端宽度预算内，联合优化三个目标：

1. **跨行对齐** — 不同行的色块边界尽可能在同一列位置
2. **尾部对齐** — 所有行以相同的总宽度结束（trailing fill 延伸最后色块的背景色）
3. **内容保留** — 对齐不应以牺牲 variable-width 内容（如 cwd 路径、git 分支）为代价

## 渲染管线

```
budgetLine (per-line)          独立裁剪每行到 maxWidth 以内
     ↓                         截断 variable-width 元素 / 隐藏低优先级元素
Extract visible + elemWidths   从 budget 后的文本计算实际显示宽度
     ↓
SmartAlign                     跨行对齐 + trailing fill + 智能缩限
     ↓                         输出: Pads (含 trailing fill) + Shrinks
BuildAlignDiagram              生成对齐前后的文本可视化（debug 用）
     ↓
Apply shrinks                  重新截断被缩限的元素文本
     ↓
Absorb padding                 将对齐 padding 回填到被截断的 variable-width 元素
     ↓                         用内容替代空白，总宽度不变，对齐不受影响
assembleWithMerge              合并同色块、注入 padding、渲染 ANSI 输出
```

## SmartAlign 内部流程

### Step 1-3: 建 boundary

对每行的可见元素做色块合并（`mergeSameColorBlocks`），累计每个色块的显示位置（包含 separator 宽度），在每个有 BgColor 的色块结束处记录一个 boundary：

```
boundary { lineIdx, blockIdx, lastElemIdx, position }
```

`position` = 从行首到该色块右边缘的累计显示列数（含 block padding + separator）。

同时收集：
- `lineWidths[li]` — 行总宽度
- `groupCounts[li]` — 有色块数量
- `lastBlockIdx[li]` — 最后一个有色块的索引

### Step 4: 排序

所有 boundary 按 position 升序排列（insertion sort，因为数据量 ≤25）。

### Step 5: 贪心循环（width-aware）

核心的 `greedyAlignWithVariance` 函数。

**外层循环按 blockIdx 分层**：G1 → G2 → G3 …，确保低层级 boundary 的 padding 落定后再评估高层级 boundary，避免后续 G1 padding 位移已对齐的 G2 boundary。

每层内部重复迭代直到无可分配 boundary：

1. **收集未分配 boundary**（仅当前 blockIdx 层级）
2. **对每个起始位置，建滑动窗口**：包含 `position` 差 ≤ `maxAlignSpread`（16 列）内的 boundary，**每行仅取一个**（防止同行不同 block 的 boundary 交叉污染）
3. **模拟 padding**：将窗口内每行的最右 boundary 对齐到 cluster target（窗口内最右的 adjusted position）。Adjusted position = `b.position + padOffset[li] - shrinkUpTo(li, b.lastElemIdx)`，其中 `shrinkUpTo` 只计算该 boundary 之前（含）元素的缩限量——boundary 之后的缩限不影响该 boundary 的视觉位置
4. **位移惩罚**：如果任何行的 padding 超过 `maxGroupPad`（6 列），effective coverage 降 1 级。这防止大间距挤压 variable-width 内容
5. **宽度检查**：如果 `maxWidth > 0` 且总宽度超限，检查 variable-width 元素的缩限容量。不够则跳过该 cluster
6. **代价评估**（四级优先级）：选出最优 cluster
7. **应用**：写入 pads、更新 padOffset/groupPads；如有溢出则执行缩限

#### 四级代价函数（P1 > P2.1 > P2.2 > P2.3）

| 层级 | 指标 | 方向 | 含义 |
|------|------|------|------|
| **P1** | effectiveLines | ↑ 越大越好 | 对齐覆盖面（参与对齐的行数），受位移惩罚降级 |
| **P2.1** | sumSqDev | ↓ 越小越好 | 所有 group 的 padding 方差（含 trailing fill），扁平计算 |
| **P2.2** | totalPad | ↓ 越小越好 | 总注入空格（含 trailing fill） |
| **P2.3** | totalShrink | ↓ 越小越好 | 总缩限量（variable-width 内容损失） |

#### 位移惩罚机制

当一个 cluster 需要给某行注入超过 `maxGroupPad`（6 列）的 padding 时，该 cluster 的 effective coverage 降 1。效果：

- 3 行 cluster + 10 列 padding → effectiveLines=2，可能输给一个只覆盖 2 行但 padding 更小的 cluster
- 2 行 cluster + 大 padding → effectiveLines=1 → 直接跳过

这防止算法为了对齐三行而在某个 group 上插入巨大间距，挤压同行的 cwd/git 等 variable-width 内容。

#### 方差计算

`computeStats` 计算所有 group 的 padding 方差，trailing fill 归属到每行最后一个有色 group：

```
对每个 group[li][bi]:
  v = groupPads[li][bi]
  如果 bi 是该行最后的有色 group:
    v += maxW - (lineWidths[li] + padOffset[li] - shrinkOffset[li])
totalPad = Σv
// 使用等比缩放避免整数除法截断：d = v*N - totalPad（N = totalGroups）
// N² 缩放因子对所有候选 cluster 相同，不影响相对排序
ssd = Σ(v * totalGroups - totalPad)²
```

`maxW` 为所有行的最大有效宽度，可选 cap 到 `maxWidth`。

#### 智能缩限

当对齐 + trailing fill 超过 `maxWidth` 时，调用 `distributeShrinkage`：

每行独立计算 overflow 并独立缩限（不存在跨行优先级）：

1. 遍历每行，如果该行有效宽度 `lineWidths[li] + padOffset[li] - shrinkOffset[li] > maxWidth`，计算 `overflow`
2. 在该行上按元素顺序查找 `shrinkRemaining[li][ei] > 0` 的元素，依次取 `min(remaining, overflow)` 进行缩限
3. 更新 `shrinkOffset[li]`、`shrinkRemaining[li][ei]`、`shrinks[li][ei]`

缩限容量 = `elemWidth - MinWidth`（仅 `IsVariable=true` 且 `MinWidth > 0` 的元素），跨迭代递减，防止元素被缩到 MinWidth 以下。

### Step 6: Trailing fill

贪心循环结束后，所有行尾部对齐到同一宽度：

```
maxTotalWidth = max(lineWidths[li] + padOffset[li] - shrinkOff[li])
if maxWidth > 0: cap at maxWidth  // R4（宽度硬约束）优先于 R2（均匀行宽）
for 每行:
  fill = maxTotalWidth - 当前行有效宽度
  加到该行最后有色块的最后元素上
```

**B 约束**：最宽行的最后一个 group 天然得到 0 trailing fill（它定义了 maxTotalWidth）。

## RenderAllLines 编排

### 1. budgetLine（不变）

独立裁剪每行。截断 variable-width 元素（`IsVariable=true`），隐藏低优先级元素。

### 2. SmartAlign

传入可见元素 + post-budget elemWidths + separator + maxWidth。返回 `AlignResult{Pads, Shrinks}`。

### 2.5. BuildAlignDiagram

在 Apply shrinks 之前，基于 SmartAlign 的原始输出生成对齐前后的文本可视化（`█`/`░` 表示色块，`·` 表示 padding）。仅用于 debug log。

### 3. Apply shrinks

对每个非零 `Shrinks[li][visIdx]`：从 post-budget 文本（非原始 buildElementText）重新截断，reserve 2 列 block padding。

### 4. Padding absorption

对齐 padding 回填到被截断的 variable-width 元素：

```
for 每个有 padding 的 variable-width 元素:
  if 当前宽度 < 原始宽度:    // 被 budgetLine 截断过
    absorb = min(padding, 原始宽度 - 当前宽度)
    从原始文本重新 Truncate 到 (当前宽度 + absorb)
    padding -= absorb
```

效果：`[ lynx/do…<gap> ]` → `[ lynx/documents/pr… ]`，总宽度不变，对齐不受影响。

### 5. Assembly

`assembleWithMerge` 合并同色元素为单个 `ColorizeBlock`，padding 注入色块内部。

## 关键常量

| 常量 | 值 | 含义 |
|------|---|------|
| `maxAlignSpread` | 16 | boundary 间最大 position 差，超过不会 cluster |
| `maxGroupPad` | 6 | 单 group 单 cluster 最大 padding，超过 effective coverage 降 1 |
| `rightReserved` | 42 | 终端宽度预留给 Claude Code 右侧通知栏的列数 |

## 硬约束

- **A (maxWidth)**：所有行 ≤ termWidth - rightReserved
- **B (零 padding 保证)**：最宽行的最后一个 group 零额外空格
- **R4 > R2**：宽度硬约束优先于行尾均匀。极窄终端下部分行可能更短

## 非 block 主题（separator 对齐模式）

当所有元素无 BgColor 时，SmartAlign 进入**非 block 模式**：

- 每个元素的右边缘都作为 boundary（而非仅有色块）
- 复用同一套贪心对齐 + level-by-level 逻辑，对齐各行的 separator 位置
- **跳过 trailing fill**：无背景色时尾部空格不可见
- `computeStats` 不计入 trailing fill 到方差计算

效果：多行的第 N 个 separator 在同一列对齐，但行尾宽度可能不同。

`assembleVisible` 在元素文本后插入对齐空格（pads）。

## 文件

- `internal/render/align.go` — SmartAlign、greedyAlignWithVariance、distributeShrinkage、assembleWithMerge、BuildAlignDiagram、PadLineToWidth
- `internal/render/renderer.go` — RenderAllLines 编排、budgetLine、padding absorption
- `internal/render/align_test.go` — SmartAlign 单元测试（~45 个）
- `internal/render/renderer_test.go` — RenderAllLines 集成测试
- `internal/render/golden_test.go` — 多行 golden 测试（testdata/multiline_block_*.golden）
- `internal/render/bench_test.go` — SmartAlign 性能基准测试
