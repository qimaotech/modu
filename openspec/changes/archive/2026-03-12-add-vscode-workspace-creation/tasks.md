## 1. 实现 workspace 文件生成

- [x] 1.1 在 `internal/engine/engine.go` 中添加 `createVSCodeWorkspace` 方法，接收 feature 名称和 feature 路径作为参数
- [x] 1.2 在 `CreateWorktree` 方法成功返回前调用 `createVSCodeWorkspace`
- [x] 1.3 workspace 文件只包含实际存在的模块、settings（Go 开发配置）、extensions（推荐扩展）

## 2. 测试

- [x] 2.1 添加单元测试验证 workspace 文件内容正确
- [x] 2.2 运行 `go test ./...` 确保测试通过

## 3. TUI 模块管理时更新 workspace

- [x] 3.1 在 TUI 中按 "m" 键添加模块后调用 `createVSCodeWorkspace` 更新 workspace
- [x] 3.2 在 TUI 中按 "m" 键删除模块后调用 `createVSCodeWorkspace` 更新 workspace
