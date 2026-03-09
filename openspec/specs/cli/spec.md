# modu CLI 命令规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md + 代码提交

## 目的

定义 modu 所有 CLI 子命令、参数及行为约定。

## 入口

- **无子命令** `modu`：进入 TUI（交互式终端时）；否则显示帮助。
- **全局 Flag**：`-c/--config` 配置文件路径（默认 `.modu.yaml`），`-o/--output` 输出格式（text | json）。

## 核心命令

| 命令 | 参数 | 说明 |
|------|------|------|
| `modu` | - | 无子命令时进入 TUI |
| `modu create` | `<feature> [--base <branch>] [--modules m1,m2]` | 并发创建基于基准分支的 worktree；可指定部分模块；feature 已存在时可继续添加模块 |
| `modu delete` | `<feature> [-f\|--force]` | 删除 worktree；默认脏检查，`--force` 跳过 |
| `modu list` | `[-v\|--verbose]` | 列出所有 worktree；verbose 显示模块、分支、状态 |
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

## 与代码的对应

- 实现：`cmd/modu/main.go`（Cobra 定义、runCreate/runDelete/runList/runInfo/runInit/runStatus/runConfigCreate/runConfigScan/runVersion）。
