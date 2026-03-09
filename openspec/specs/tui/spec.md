# modu TUI 规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

基于 Bubble Tea 的状态机 TUI，提供 worktree 列表查看、创建/删除 worktree，删除前二次确认。

## 入口

- 裸命令 `modu` 且在交互式终端时进入 TUI。
- 命令 `modu tui` 显式启动 TUI；无配置文件时可启动配置向导。

## 状态机

1. **LoadingState**：并发执行 init 或 create 时，显示每个 Module 的当前进度（如 `api-server: Cloning...`）。
2. **ListState**：展示所有 feature 列表；光标选中时展示该环境下各模块的 Branch 与 Status（Clean/Dirty）。
3. **ConfirmState**：删除前的二次确认。
4. **ErrorState**：操作失败时显示错误详情，允许重试。

## UI 表现

- 多行并行进度（Multi-Spinner），实时显示各模块任务状态。
- 键盘：上下选择、回车确认、ESC 取消。

## 能力范围

- **只读**：worktree 列表、分支、模块状态。
- **写**：创建 worktree、删除 worktree（删除前必须确认；脏检查由 engine 执行，TUI 展示结果）。

## 与代码的对应

- 实现：`internal/ui`（Bubble Tea 模型、状态转换）；`internal/ui/config_wizard.go`（配置向导）。
