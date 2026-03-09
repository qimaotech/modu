# Tasks: update-feature-and-cli

## 1. Engine 层

- [x] 1.1 在 `internal/engine/engine.go` 新增 `UpdateWorktree(ctx context.Context, feature string) (success int, failed map[string]error)`，对 `worktreeRoot/feature` 及该目录下存在的模块路径并发执行 fetch + rebase
- [x] 1.2 为 `UpdateWorktree` 编写单元测试：成功场景、部分失败场景

## 2. CLI 子命令

- [x] 2.1 在 `cmd/modu/main.go` 新增 `update` 子命令，可选参数 `<feature>`
- [x] 2.2 无参数时调用 `UpdateMainProject`，有参数时校验 feature 存在后调用 `UpdateWorktree`，并输出成功/失败摘要

## 3. TUI feature 更新

- [x] 3.1 在 feature 操作菜单中增加“更新代码 (u)”选项
- [x] 3.2 菜单选中“更新代码”或按 u 时，调用 `executeUpdateWorktree`（内部调用 `Engine.UpdateWorktree`），进入 loading 后显示结果并刷新列表
- [x] 3.3 在列表视图中选中 feature 时，u 键触发该 feature 的更新（与主项目 u 键逻辑区分：主项目调用 `executeUpdateCode`，feature 调用 `executeUpdateWorktree`）

## 4. 测试与验证

- [x] 4.1 单元测试：`UpdateWorktree` 通过
- [x] 4.2 手动验证：`modu update`、`modu update <feature>` 行为符合 spec
- [x] 4.3 手动验证：TUI 中 feature 菜单与 u 键更新行为符合 spec
