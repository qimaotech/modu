# Proposal: update-feature-and-cli

## Why

当前仅支持在 TUI 中更新**主项目**（workspace 根仓库 + 所有模块）。用户也需要对**单个 feature worktree** 执行更新（该 feature 下的主项目 worktree + 该 feature 下的各模块 worktree），并在命令行中通过子命令统一执行“更新”操作，便于脚本化与 CI。

## What Changes

1. **Feature worktree 支持更新**：在 TUI 中选中某个 feature 时，支持通过快捷键或菜单执行“更新代码”，对该 feature 的 worktree（主项目 + 该 feature 下的所有模块）执行 git fetch + rebase。
2. **新增 CLI 子命令 `modu update`**：
   - `modu update`：更新主项目（与当前 TUI 主项目更新行为一致，即 workspace + 所有模块）。
   - `modu update <feature>`：更新指定 feature 的 worktree（该 feature 目录下的主项目 + 其下所有模块）。
3. **Engine 层**：新增 `UpdateWorktree(ctx, feature string) (success int, failed map[string]error)`，对指定 feature 的 worktree 并发执行 fetch + rebase。

## Capabilities

### New Capabilities

- **cli-update**：CLI 子命令 `modu update [feature]`，无参数时更新主项目，带 feature 时更新该 feature 的 worktree。
- **tui-feature-update**：TUI 中 feature 条目支持“更新代码”（u 键/菜单），对当前选中的 feature worktree 执行 fetch + rebase。

### Modified Capabilities

- （无：主项目更新与现有 tui-main-project 行为一致，仅扩展“更新”到 feature 与 CLI。）

## Impact

- 修改 `internal/engine/engine.go`：新增 `UpdateWorktree(ctx, feature string)`。
- 修改 `internal/ui/ui.go`：feature 菜单/列表下支持 u 与“更新代码”，调用 `UpdateWorktree`。
- 修改 `cmd/modu/main.go`：新增 `update` 子命令，根据是否有参数调用 `UpdateMainProject` 或 `UpdateWorktree`。
