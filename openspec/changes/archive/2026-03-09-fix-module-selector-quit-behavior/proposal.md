## Why

在交互式模块选择界面中，用户按 `q` 或 `ctrl+c` 退出时，当前实现会误判为"用户想删除已存在的模块"，导致意外删除已创建的模块。正确的行为应该是：用户主动退出时保留所有已存在的模块，不执行任何操作。

## What Changes

- 修改 `ui.SelectModules` 函数，增加返回用户是否主动退出（按 q/ctrl+c）的状态标识
- 修改 `cmd/modu/main.go` 中的交互逻辑：检测到用户主动退出时，直接 return 保留模块，而不是执行删除
- 在删除模式的提示文案中增加"如需保留请按 q 退出"的说明

## Capabilities

### New Capabilities
- 无新能力引入

### Modified Capabilities
- 无现有能力需求变更

## Impact

- 修改文件：
  - `internal/ui/ui.go` - `SelectModules` 函数签名和返回值
  - `cmd/modu/main.go` - 交互逻辑处理
