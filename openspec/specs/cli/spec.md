# modu CLI 命令规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md + 代码提交

## Purpose

定义 modu 所有 CLI 子命令、参数及行为约定，包括命令入口、参数形态、输出协议、错误提示，以及 create/delete/list/update 等用户可见行为。

## 入口

- **无子命令** `modu`：进入 TUI（交互式终端时）；否则显示帮助。
- **全局 Flag**：`-c/--config` 配置文件路径（默认 `.modu.yaml`），`-o/--output` 输出格式（text | json）。

## 核心命令

| 命令 | 参数 | 说明 |
|------|------|------|
| `modu` | - | 无子命令时进入 TUI |
| `modu create` | `<feature> [--base <branch>] [--modules m1,m2] [--no-fetch]` | 并发创建基于基准分支的 worktree；可指定部分模块；feature 已存在时可继续添加模块；`--no-fetch` 基于本地分支创建 |
| `modu delete` | `<feature> [-f\|--force]` | 删除 worktree；默认脏检查，`--force` 跳过 |
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

### Requirement: CLI create supports explicit local-only mode

`modu create` SHALL support an explicit parameter for feature creation from local branch state without pre-create fetch.

#### Scenario: Local-only parameter skips fetch

- **WHEN** the user runs `modu create <feature> --no-fetch`
- **THEN** the CLI creates the feature from local branch state without pre-create fetch

#### Scenario: Fetch failure keeps original error path with hint

- **WHEN** the user runs `modu create <feature>` and pre-create fetch fails
- **THEN** the CLI reports the original create error
- **AND** the CLI prints a Chinese hint telling the user they can rerun with `--no-fetch` to skip remote fetch and create from local code

#### Scenario: Fetch failure does not prompt in CLI

- **WHEN** the user runs `modu create <feature>` in an interactive terminal and pre-create fetch fails
- **THEN** the CLI does not ask for confirmation and does not continue from local branch state in the same command

#### Scenario: Auto-fetch disabled behaves like local-only mode

- **WHEN** `modu create <feature>` runs with `auto-fetch: false`
- **THEN** the CLI creates the feature from local branch state without pre-create fetch

## 与代码的对应

- 实现：`cmd/modu/main.go`（Cobra 定义、runCreate/runDelete/runList/runInfo/runInit/runStatus/runConfigCreate/runConfigScan/runVersion）。
