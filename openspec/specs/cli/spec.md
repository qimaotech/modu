# modu CLI 命令规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md + 代码提交

## Purpose

定义 modu 所有 CLI 子命令、参数及行为约定，确保命令入口、配置默认值、分支基准和错误输出在不同执行路径中保持一致。

## 入口

- **无子命令** `modu`：进入 TUI（交互式终端时）；否则显示帮助。
- **全局 Flag**：`-c/--config` 配置文件路径（默认 `.modu.yaml`），`-o/--output` 输出格式（text | json）。

## 核心命令

| 命令 | 参数 | 说明 |
|------|------|------|
| `modu` | - | 无子命令时进入 TUI |
| `modu create` | `<feature> [--base <branch>] [--modules m1,m2]` | 并发创建基于基准分支的 worktree；可指定部分模块；feature 已存在时可继续添加模块 |
| `modu delete` | `<feature> [-f\|--force] [--allow-unpushed]` | 删除 worktree；默认脏检查，`--force` 跳过脏检查，`--allow-unpushed` 显式允许删除未推送本地分支 |
| `modu list` | `[-v\|--verbose] [-a\|--all]` | 列出所有 worktree；verbose 显示模块、分支、状态；-a 显示主项目及模块分支 |
| `modu info` | `<feature>` | 查看指定 feature 详情 |
| `modu init` | `[--scan]` | 并发克隆配置中的仓库；`--scan` 可先扫描发现模块再初始化 |
| `modu status` | - | 批量展示所有模块 Dirty 状态 |
| `modu version` | - | 显示版本信息（来自 git describe/commit/date） |

## 配置相关命令

| 命令 | 参数 | 说明 |
|------|------|------|
| `modu config create` | `[--workspace] [--worktree-root] [--default-base] [--module name=url...]` | 创建配置文件 |
| `modu config scan` | - | 扫描当前/workspace 发现模块，可导出或更新配置 |
| `modu tui` | - | 显式启动 TUI；无配置时可启动配置向导 |

## 行为约定

- 子命令存在时始终走 CLI，不走 TUI。
- JSON 输出（`-o json`）时，成功/失败结构需符合 [errors 规范](./../errors/spec.md) 中的机器输出协议。

## Requirements

### Requirement: Delete protects unpushed local branches
The `modu delete` command SHALL check whether local branches that would be deleted are fully pushed before removing worktrees or local branches.

#### Scenario: All branches are pushed
- **WHEN** the user runs `modu delete <feature>` and every branch that would be deleted has no local commits ahead of its remote branch
- **THEN** deletion proceeds without an additional unpushed-branch confirmation

#### Scenario: Delete has unpushed branches without explicit allow flag
- **WHEN** the user runs `modu delete <feature>` and at least one branch that would be deleted is not fully pushed
- **THEN** the command fails with `ERR_UNPUSHED_BRANCH` without deleting worktrees or local branches
- **THEN** the error message tells the user to rerun with `--allow-unpushed` to confirm deletion

#### Scenario: Delete has unpushed branches with explicit allow flag
- **WHEN** the user runs `modu delete <feature> --allow-unpushed` and at least one branch that would be deleted is not fully pushed
- **THEN** deletion proceeds without interactive confirmation

#### Scenario: Force does not bypass unpushed protection
- **WHEN** the user runs `modu delete <feature> --force` and at least one branch that would be deleted is not fully pushed
- **THEN** the command fails with `ERR_UNPUSHED_BRANCH` unless `--allow-unpushed` is also provided

### Requirement: Create uses configured default base
The CLI SHALL use the loaded configuration `default-base` as the base branch for `modu create <feature>` when the user does not provide `--base`.

#### Scenario: Create without base flag
- **WHEN** the user runs `modu create feature/foo` and `.modu.yaml` contains `default-base: release/1.2`
- **THEN** the system creates the main project worktree from `release/1.2`
- **AND** modules without module-level `base-branch` are created from `release/1.2`

#### Scenario: Create with explicit base flag
- **WHEN** the user runs `modu create feature/foo --base main` and `.modu.yaml` contains `default-base: develop`
- **THEN** the system creates the main project worktree from `main`
- **AND** modules without module-level `base-branch` are created from `main`

#### Scenario: Module base branch still wins
- **WHEN** the user runs `modu create feature/foo --base main`
- **AND** a selected module has `base-branch: release/module`
- **THEN** that module worktree is created from `release/module`

## 与代码的对应

- 实现：`cmd/modu/main.go`（Cobra 定义、runCreate/runDelete/runList/runInfo/runInit/runStatus/runConfigCreate/runConfigScan/runVersion）。
