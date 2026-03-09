# Tasks: tui-main-project

## 1. Engine 层实现

- [x] 1.1 在 `internal/engine/engine.go` 新增 `MainProjectStatus` 结构体，包含 Name、Path、IsDirty、Branch 字段
- [x] 1.2 在 `Engine` 新增 `GetMainProject(ctx context.Context) (*MainProjectStatus, error)` 方法
- [x] 1.3 在 `Engine` 新增 `UpdateMainProject(ctx context.Context) (success int, failed map[string]error)` 方法，实现并发 fetch + rebase

## 2. TUI 数据结构

- [x] 2.1 在 `internal/ui/ui.go` 新增 `MainProjectEntry` 类型，实现 `ListEntry` 接口
- [x] 2.2 在 `App` 结构体新增 `mainProject *MainProjectStatus` 字段
- [x] 2.3 修改 `loadEnvs` 方法，同时加载主项目信息

## 3. TUI 列表视图

- [x] 3.1 修改 `renderList` 方法，在顶部渲染主项目条目
- [x] 3.2 主项目条目格式："→ <主项目名> [主项目] (<dirty状态>) [<分支名>]"
- [x] 3.3 修改 `handleListKey` 方法，处理主项目条目的导航（上下键在主项目和 feature 间切换）

## 4. TUI 主项目菜单

- [x] 4.1 在 `App` 新增 `isMainProjectMenu bool` 字段区分菜单类型
- [x] 4.2 修改 `handleMenuKey` 方法，根据选中项类型执行不同逻辑
- [x] 4.3 修改 `renderMenu` 方法，根据 isMainProjectMenu 渲染不同菜单
- [x] 4.4 实现主项目菜单快捷键：o 打开 VS Code，u 更新代码

## 5. TUI 更新代码功能

- [x] 5.1 在 `App` 新增 `executeUpdateCode` 方法，调用 Engine.UpdateMainProject
- [x] 5.2 修改 `handleListKey` 方法，添加 u 快捷键处理（主项目有效，Feature 无效）
- [x] 5.3 修改 `handleMenuKey` 方法，添加 u 快捷键处理（主项目菜单有效）
- [x] 5.4 在 loading 完成后正确刷新列表

## 6. 测试

- [x] 6.1 单元测试：Engine.UpdateMainProject 成功场景
- [x] 6.2 单元测试：Engine.UpdateMainProject 部分失败场景
- [x] 6.3 手动测试：TUI 列表正确显示主项目
- [x] 6.4 手动测试：主项目菜单快捷键功能正常
