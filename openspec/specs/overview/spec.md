# modu 概述规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

modu 是基于 Go 的多模块 Git Worktree 管理工具，替代复杂 Shell 脚本，通过 TUI 和强类型 CLI 提供跨平台一致、安全、高效的工作流管理。

## 需求

### 解决的问题

- 团队成员多数无 Shell 经验
- Mac/Linux 行为可能不一致
- 分发给团队需安装 task runner

### 目标用户

- **主要**: 团队成员
- **次要**: 大模型/脚本调用（通过 `-o json`）

### 环境约束

| 约束 | 要求 |
|------|------|
| Go | 1.21+（利用 `errors.Join` 处理并发多错） |
| Git | 2.25+（支持 worktree 核心功能） |
| OS | Linux、macOS (Darwin) 完全兼容；Windows 仅支持 WSL2 |

## 非目标

- 不替代 Git 本身，仅编排多仓库 worktree 工作流
- 不提供通用 task runner，仅围绕 worktree 的 create/list/delete/status
