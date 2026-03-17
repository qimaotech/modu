## Context

当前 TUI 程序使用 lipgloss 库构建 UI，样式定义在 `internal/ui/ui.go` 中使用硬编码的 ANSI 颜色值（如 `Color("252")` 表示浅灰）。这些颜色设计时针对深色终端背景，在浅色背景下可读性极差。

## Goals / Non-Goals

**Goals:**
- 使用 lipgloss 的 `AdaptiveColor` 实现深浅主题自动适配
- 深色模式保持现有配色（黑底+浅色文字）
- 浅色模式使用高对比度配色（白底+深色文字）
- 仅修改 UI 样式定义，不涉及业务逻辑变更

**Non-Goals:**
- 不添加用户手动切换主题的功能（auto 即可）
- 不修改其他模块的样式（仅 TUI）
- 不引入新的外部依赖

## Decisions

### 方案选择：lipgloss AdaptiveColor vs 环境变量检测

**选用方案：lipgloss AdaptiveColor**

- **理由**：lipgloss 内置支持 `AdaptiveColor`，可以同时定义深色/浅色两套颜色，运行时根据终端背景自动选择，无需额外代码
- **备选**：通过读取环境变量 `COLORFGBG` 或调用系统 API 检测，需要额外代码且不够精确

### 配色方案：柔和渐变风格

| 样式 | 深色模式 | 浅色模式 |
|------|---------|---------|
| headerStyle 背景 | `236` (深灰) | `254` (浅灰) |
| headerStyle 前景 | `86` (浅青) | `25` (深蓝) |
| itemStyle 前景 | `252` (浅灰) | `245` (深灰) |
| selectedItemStyle 前景 | `86` (浅青) | `21` (蓝色) |
| errorStyle 前景 | `196` (红色) | `196` (红色) |
| successStyle 前景 | `82` (绿色) | `28` (深绿) |

**理由**：
- 浅色模式下使用深灰前景色（245）替代浅灰（252），确保对比度
- 保留 error/success 颜色在两种模式下的一致性
- 选用的颜色在 256 色 ANSI 调色板中对比度良好

## Risks / Trade-offs

- **兼容性风险**：部分终端可能不完全支持 lipgloss 的自适应颜色检测
  - →  Mitigation：lipgloss 在大多数主流终端（iTerm2, Terminal.app, VSCode Terminal）中工作良好，作为可选依赖无需额外处理
- **用户体验**：浅色模式下的配色可能不符合所有用户审美
  - → Mitigation：后续可添加配置项允许用户覆盖自动检测结果（当前保持简单，YAGNI）

## Migration Plan

1. 修改 `internal/ui/ui.go` 中的全局样式变量
2. 运行单元测试验证无回归
3. 手动在深/浅两种终端背景下测试确认效果

**回滚方案**：直接回滚代码修改，无数据迁移需求
