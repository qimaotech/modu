## Context

当前 `modu list` 命令的输出格式固定显示每个模块的 clean/dirty 状态。用户希望能控制这个信息的显示。

## Goals / Non-Goals

**Goals:**
- 添加 `-s/--status` flag 到 `modu list` 命令
- 带 `-s` 时显示状态，不带时不显示

**Non-Goals:**
- 不修改其他命令的行为
- 不添加新的 capability

## Decisions

1. **使用 `-s` 简写**: 与其他命令（如 `-v`）保持一致的简洁风格

2. **默认不显示状态**: 保持向后兼容，默认输出更简洁

3. **实现位置**:
   - Flag 定义在 `cmd/modu/main.go`
   - 输出格式化在 `internal/output/output.go`

## Risks / Trade-offs

无显著风险，这是一个简单的 CLI 修改。
