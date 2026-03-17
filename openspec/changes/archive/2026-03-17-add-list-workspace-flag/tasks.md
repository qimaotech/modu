## 1. Engine 层

- [x] 1.1 在 `engine.go` 中添加 `GetMainProjectModules` 方法，获取主项目及其模块的分支状态

## 2. CLI 层

- [x] 2.1 在 `cmd/modu/main.go` 的 `list` 命令中添加 `-a` / `--all` flag
- [x] 2.2 修改 `runList` 函数，当 `-a` 为 true 时获取主项目模块信息

## 3. Output 层

- [x] 3.1 在 `output.go` 中添加 `MainProjectInfo` 结构体
- [x] 3.2 添加 `FormatMainProjectResponse` 方法处理文本和 JSON 格式输出

## 4. 测试

- [x] 4.1 添加 `TestFormatMainProjectResponse` 单元测试
