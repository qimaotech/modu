# modu 核心引擎规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

定义 create/delete/list 等核心流程、并发策略、事务性创建与脏检查算法。

## 引擎职责

- 并发调度：使用 `golang.org/x/sync/errgroup`，限流为 `Config.Concurrency`。
- 事务性创建：Create 失败时回滚已创建的 worktree 目录。
- 脏检查：Delete 前（除非 `--force`）检查各模块是否存在未提交修改。
- **模块认定**：凡涉及「模块」的 list/create/delete 逻辑，仅处理 `Config.Modules` 中的目录；feature 下的其他子目录（如 `.claude`、`openspec` 等）一律忽略，不展示、不参与增删、不参与脏检查。

## CreateWorktree（事务性并发创建）

1. **Pre-check**：若 `worktree-root/<feature>` 已存在，可支持“继续添加模块”或报错（由 CLI 层选择）；主项目 worktree 位于 feature 目录根，子模块位于 `feature/<module.Name>`。
2. **Execution**：主项目先创建 worktree，再 errgroup 并发为各模块执行 `git fetch` + `git worktree add`；已存在的模块目录跳过。
3. **Rollback**：若任一模块失败，收集已成功创建的路径，依次 `RemoveWorktreeAndBranch(ctx, repoPath, path, dirName)` + `os.RemoveAll`，再删除主项目 worktree 与 feature 目录；其中 `dirName` 为 `featureToDirName(feature)`，返回 `ERR_PARTIAL_FAILURE` 或等价错误。

## DeleteWorktree

1. 若 feature 目录不存在，返回 `ERR_FEATURE_NOT_FOUND`。
2. 若未使用 `--force` 且 `Config.StrictDirty` 为 true：构建 `WorktreeEnv` 时仅包含 **配置内模块** 子目录，调用 **CheckDirty**；若有 dirty 模块，返回 `ERR_DIRTY_WORKTREE`。
3. 仅对 **配置内模块** 依次执行 `RemoveWorktreeAndBranch(ctx, repoPath, modulePath, dirName)`，再对主项目执行 `RemoveWorktreeAndBranch(ctx, workspace, mainProjectPath, dirName)`，最后 `os.RemoveAll(featurePath)`。`dirName` 为 `featureToDirName(feature)`，供 gitproxy 校验「当前分支 slug」与目录名一致后再删分支。不对非配置目录单独调用 RemoveWorktreeAndBranch。

## CheckDirty

- 输入：`WorktreeEnv`（含各模块 Path）。
- 对每个模块调用 gitproxy `GetStatus(path)`；若 `Status.IsDirty` 则加入结果列表。
- 返回：dirty 的 `[]ModuleStatus`。

## ListWorktrees

- 扫描 `worktree-root` 下子目录，每个子目录名视为 feature 名。
- 对每个 feature 仅收集 **配置内模块**（名称在 `Config.Modules` 中且存在的子目录），构造 `WorktreeEnv`（Name、Base、Modules）；模块的 Branch/IsDirty 通过 gitproxy GetStatus 获取。非配置目录（如 `.claude`、`openspec`）不列入 Modules。

## 命令与 Git 原语映射

| modu 命令 | 内部逻辑 |
|-----------|----------|
| init | `git clone <url> <workspace>/<name>`（已存在则跳过） |
| create &lt;f&gt; | 主项目：`git -C <workspace> fetch` + `worktree add <wt-root>/<f> -b <f> <base>`；模块：`git -C <workspace>/<name> fetch` + `worktree add <wt-root>/<f>/<name> -b <f> <base>` |
| list | 扫描 `worktree-root`，对每个 feature 读目录 + GetStatus |
| delete &lt;f&gt; | Dirty Check（可选）→ `git worktree remove` / RemoveWorktreeAndBranch → `rm -rf <wt-root>/<f>` |

## 与代码的对应

- 实现：`internal/engine/engine.go`（Init、CreateWorktree、DeleteWorktree、CheckDirty、ListWorktrees）。
