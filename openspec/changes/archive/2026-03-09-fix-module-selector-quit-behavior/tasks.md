## 1. 代码修改

- [x] 1.1 修改 `ui.SelectModules` 函数签名，增加返回 `isQuit bool`
- [x] 1.2 修改 `ModuleSelector` 使用 `quitting` 字段返回退出状态
- [x] 1.3 修改 `main.go` 中 `runCreate` 函数：检测 `isQuit` 为 true 时直接 return

## 2. 提示文案优化

- [x] 2.1 在删除模式提示文案中增加"如需保留请按 q 退出"

## 3. 验证

- [x] 3.1 编译验证：`go build ./...`
