## Context

当前 TUI 应用 `internal/ui/ui.go` 使用 Bubble Tea 框架实现，支持列出 worktree、删除、打开 VS Code 等功能。现有交互：
- 列表视图：上下键选择，Enter 进入删除确认，o 打开 VS Code，q 退出
- 删除确认：y/Enter 确认，n/esc 取消

## Goals / Non-Goals

**Goals:**
- 在列表视图新增操作菜单，按 Enter 进入
- 操作菜单支持打开 VS Code 和删除两个选项
- 支持在列表视图直接按 d 触发删除确认（仍需二次确认）
- 操作菜单按 Enter 可执行当前选中操作
- 打开 VS Code 后自动返回列表视图
- 保持现有功能和交互习惯

**Non-Goals:**
- 不修改删除二次确认流程
- 不添加新的操作类型（仅删除和打开 VS Code）

## Decisions

1. **新增 "menu" 状态**
   - 在 App 结构的 state 字段添加 "menu" 状态
   - menu 状态下渲染操作菜单项列表

2. **操作菜单使用独立的光标选择**
   - 新增 `menuSelected` 字段跟踪菜单选中项
   - 0 = 打开 VS Code，1 = 删除（打开在前，删除在后）

3. **保留列表视图直接按 o 的功能**
   - 不改变现有直接打开 VS Code 的快捷键
   - 操作菜单提供另一种进入方式

4. **Enter 行为变更**
   - 原行为：Enter → 删除确认
   - 新行为：Enter → 操作菜单

5. **Enter 执行选中操作**
   - 在操作菜单按 Enter 执行当前选中的操作
   - 打开 VS Code 后自动返回列表视图

6. **打开 VS Code 后自动返回**
   - 按 o 或 Enter 执行"打开 VS Code"后，返回列表视图
   - 提供更流畅的用户体验

## Risks / Trade-offs

- [低风险] 用户习惯变更：Enter 不再直接进入删除确认
  - 缓解：提供 d 快捷键直接删除，功能等价
- [低风险] 状态管理复杂度增加
  - 缓解：新状态与现有状态分离，逻辑清晰
