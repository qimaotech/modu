## Context

当前 modu TUI 在列表视图中，选中 feature 后按 Enter 会进入删除确认流程。用户需要手动在 Finder 或终端中打开对应的项目目录。

## Goals / Non-Goals

**Goals:**
- 在 TUI 列表视图添加按 `o` 键打开 VS Code 的功能
- 打开对应 feature 的主项目目录

**Non-Goals:**
- 不修改 TUI 的其他交互逻辑
- 不添加其他编辑器支持（仅支持 VS Code）

## Decisions

1. **使用 `code` 命令打开项目**
   - 替代方案：使用 `open` 命令打开 Finder
   - 选择理由：用户明确要求使用 VS Code 的 `code` 命令

2. **使用异步调用 `code` 命令**
   - 替代方案：同步等待命令执行
   - 选择理由：VS Code 启动后无需等待，TUI 保持响应

3. **主项目路径获取**
   - 从 `WorktreeEnv.MainProject.Path` 获取主项目路径
   - 如果 `MainProject` 为 nil，显示错误

4. **保持 TUI 运行**
   - 替代方案：打开后退出 TUI
   - 选择理由：用户可能需要继续操作其他 feature

## Risks / Trade-offs

- [低风险] `code` 命令不存在时：显示错误提示，不影响 TUI 继续运行
