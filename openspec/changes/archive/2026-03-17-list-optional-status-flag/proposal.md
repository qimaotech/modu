## Why

当前 `modu list` 命令总是显示每个模块的 clean/dirty 状态，但在某些场景下用户不需要这个信息，状态显示会增加输出噪音。添加可选的状态显示 flag 让输出更灵活。

## What Changes

- 为 `modu list` 命令添加 `-s/--status` flag
- 带 `-s` flag 时显示状态（如 `(clean)`、`(dirty)`）
- 不带 flag 时不显示状态
- 默认行为不变（不显示状态）

## Capabilities

### New Capabilities
无

### Modified Capabilities
无（这是 CLI 行为的简单修改，不涉及 spec 级别的变化）

## Impact

- **修改文件**:
  - `cmd/modu/main.go`: 添加 `-s/--status` flag
  - `internal/output/output.go`: 修改 `FormatListResponse` 支持可选状态显示
