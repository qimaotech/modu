## Context

当前 `modu create` 命令在交互式选择模块时，如果 feature 已存在已创建的模块，会让用户选择要保留哪些模块。但用户按 `q` 或 `ctrl+c` 退出时，代码没有区分"用户主动退出"和"用户按回车确认但没选模块"这两种情况，统一当作后者处理，导致误删除模块。

涉及的代码文件：
- `internal/ui/ui.go` - `SelectModules` 函数和 `ModuleSelector` 结构
- `cmd/modu/main.go` - `runCreate` 函数中的模块处理逻辑

## Goals / Non-Goals

**Goals:**
- 区分用户按 `q`/`ctrl+c` 主动退出和按回车确认但未选择模块两种行为
- 用户主动退出时保留所有已存在的模块，不执行任何操作
- 用户按回车但未选模块时，保持原有的删除模式，并增加提示文案

**Non-Goals:**
- 不修改非交互模式下的行为
- 不修改模块选择的 UI 交互细节（光标移动、选中状态等）

## Decisions

1. **在 `SelectModules` 返回值中增加退出状态标识**
   - 方案：在返回值中增加 `bool` 类型标识用户是否按 `q`/`ctrl+c` 退出
   - 理由：最小改动，只需修改函数签名和返回值处理

2. **在 main.go 中根据退出状态决定行为**
   - `isQuit == true` → 直接 return，保留模块
   - `len(selectedModules) == 0` → 删除模式，带提示文案

## Risks / Trade-offs

- **风险低**：改动范围小，仅涉及两处代码修改
- 无需数据迁移或回滚策略
