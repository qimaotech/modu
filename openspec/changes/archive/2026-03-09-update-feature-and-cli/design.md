# Design: update-feature-and-cli

## Context

- 主项目更新已实现：`Engine.UpdateMainProject(ctx)` 对 workspace 根目录及配置中的各模块路径执行 fetch + rebase，TUI 主项目菜单/`u` 键可触发。
- Feature worktree 结构：每个 feature 对应 `worktreeRoot/<feature>` 目录，该目录即主项目的 worktree，其下子目录为各模块的 worktree（与 `Config.Modules` 对应）。
- 需扩展：对**单个 feature** 的 worktree（主项目 + 该 feature 下存在的模块）做同样“更新”操作，并暴露为 CLI 子命令。

## Goals / Non-Goals

**Goals:**
- Engine 新增 `UpdateWorktree(ctx, feature string) (success int, failed map[string]error)`，对指定 feature 的 worktree 并发 fetch + rebase。
- TUI 中选中 feature 时，`u` 键与菜单“更新代码”调用 `UpdateWorktree`，交互与主项目更新一致（loading → 结果消息 → 刷新列表）。
- CLI 新增 `modu update`、`modu update <feature>`，分别调用 `UpdateMainProject`、`UpdateWorktree`，输出成功/失败摘要。

**Non-Goals:**
- 不改变现有 `UpdateMainProject` 语义；不支持“只更新部分模块”的粒度。
- 不新增配置项；沿用现有 `Concurrency` 等配置。

## Decisions

**1. Engine：UpdateWorktree 与 UpdateMainProject 复用同一 Rebase 语义**

- `UpdateWorktree(ctx, feature string)`：`featurePath := filepath.Join(WorktreeRoot, feature)`；对 `featurePath` 执行一次 Rebase（主项目 worktree）；对 `featurePath/<module.Name>` 仅当目录存在时执行 Rebase（与 ListWorktrees 中“存在的模块”一致）。
- 并发与错误汇总方式与 `UpdateMainProject` 一致：errgroup、`failed map[string]error`、success 计数（主项目 1 + 实际存在的模块数中未失败者）。
- **理由**：逻辑与主项目更新对称，便于维护和测试。

**2. CLI：`modu update` 无参数 = 主项目，有参数 = feature**

- 子命令：`update`，可选参数 `<feature>`。
- `modu update` → 调用 `UpdateMainProject`；`modu update <feature>` → 校验 feature 存在（如 ListWorktrees 或目录存在），再调用 `UpdateWorktree(ctx, feature)`。
- 输出：成功时打印“更新成功: 主项目”或“更新成功: 主项目 + N 个模块”或“更新成功: feature &lt;name&gt;（主项目 + N 个模块）”；有失败时打印失败数量与名称列表（与 TUI 文案风格一致）。
- **理由**：一个子命令覆盖两种场景，符合“对项目进行更新”的直觉。

**3. TUI：feature 菜单增加“更新代码”，列表 u 键对 feature 有效**

- Feature 菜单与主项目菜单对齐：增加“更新代码 (u)”选项；选中 feature 时在列表按 `u` 同样触发更新。
- 触发后：进入 loading，执行 `UpdateWorktree(ctx, selectedFeature)`，完成后刷新列表并显示结果消息（与主项目更新相同流程）。
- **理由**：主项目与 feature 的“更新”交互一致，减少认知负担。

## Risks / Trade-offs

- **风险**：feature 目录存在但非合法 worktree（如被手动改坏），Rebase 可能报错。**缓解**：按现有逻辑记录到 `failed`，不中断其他仓库；TUI/CLI 展示失败信息。
- **取舍**：不区分“主项目 worktree”与“模块 worktree”的更新顺序；并发执行。与 `UpdateMainProject` 一致，简单可预期。
