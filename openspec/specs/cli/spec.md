# modu CLI 命令规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md + 代码提交

## Purpose

定义 modu 所有 CLI 子命令、参数、输出格式和命令启动行为约定，确保命令入口在脚本与交互使用中保持一致。

## 入口

- **无子命令** `modu`：进入 TUI（交互式终端时）；否则显示帮助。
- **全局 Flag**：`-c/--config` 配置文件路径（默认 `.modu.yaml`），`-o/--output` 输出格式（text | json）。

## 核心命令

| 命令 | 参数 | 说明 |
|------|------|------|
| `modu` | - | 无子命令时进入 TUI |
| `modu create` | `<feature> [--base <branch>] [--modules m1,m2]` | 并发创建基于基准分支的 worktree；可指定部分模块；feature 已存在时可继续添加模块 |
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
### Requirement: CLI runs delete backup cleanup after config load
CLI commands that successfully load an existing `.modu.yaml` via normal or scan config loading SHALL invoke delete backup cleanup once before executing their main command behavior. Cleanup failure SHALL be reported as a warning on stderr and MUST NOT block the requested command or corrupt JSON stdout.

#### Scenario: Configured command starts
- **WHEN** `modu create`, `modu delete`, `modu default-select`, `modu list`, `modu info`, `modu status`, `modu update`, `modu init`, or `modu config scan` successfully loads existing configuration
- **THEN** the command invokes delete backup cleanup once before its main behavior

#### Scenario: Cleanup warning does not block command
- **WHEN** startup cleanup fails after configuration loading succeeds
- **THEN** the CLI prints a warning to stderr and continues with the requested command

#### Scenario: Cleanup warning preserves JSON stdout
- **WHEN** startup cleanup fails and the requested command uses `-o json`
- **THEN** the cleanup warning is not written to stdout

#### Scenario: Command without existing config load does not clean
- **WHEN** `modu version`, help output, or `modu config create` runs without loading an existing config
- **THEN** delete backup cleanup is not required to run

### Requirement: CLI delete exposes backup path
`modu delete <feature>` SHALL report the created backup archive path when deletion succeeds.

#### Scenario: Text delete output includes backup path
- **WHEN** `modu delete my-feature` succeeds with text output
- **THEN** the output includes the feature name and backup archive path

#### Scenario: JSON delete output includes backup path
- **WHEN** `modu delete my-feature -o json` succeeds
- **THEN** the JSON response includes `backupPath`

## 与代码的对应

- 实现：`cmd/modu/main.go`（Cobra 定义、runCreate/runDelete/runList/runInfo/runInit/runStatus/runConfigCreate/runConfigScan/runVersion）。
