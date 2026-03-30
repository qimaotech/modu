## Why

当前 `modu info <feature>` 要求必须传入 feature 参数。但用户在 worktree 目录内工作时，往往已经在某个 feature 下，希望直接查看当前 feature 的状态而无需记忆并输入 feature 名称。这与 `git status` 在仓库内自动感知当前分支的行为一致。

## What Changes

- **`modu info` 无参数时**：自动从当前工作目录向上回溯，找到 `worktreeRoot` 下一级的目录名作为 feature，调用 `GetWorktreeInfo` 显示该 feature 的详情。
- **参数仍保留**：传入 `<feature>` 参数时行为不变。
- **边界处理**：当前目录不在任何 feature 下、或恰好在 `worktreeRoot` 本身时，给出明确错误提示。

## Capabilities

### New Capabilities

- `cli-info-autodetect`: `modu info` 无参数时自动推断当前所属 feature 并展示详情。

### Modified Capabilities

- `cli`: `modu info` 命令的参数约束从必填改为可选（MaximumNArgs(1)），行为参见 `cli-info-autodetect`。

## Impact

- 修改 `cmd/modu/main.go` 中 `info` 命令的 `Args` 约束及 `runInfo` 函数。
- 新增 engine 层辅助函数 `GetCurrentFeature(ctx, cwd) (feature string, err error)`（或内联于 `runInfo`）。
