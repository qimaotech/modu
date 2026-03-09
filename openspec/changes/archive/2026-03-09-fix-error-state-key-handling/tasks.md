## 1. 修改 error 状态按键处理

- [x] 1.1 在 `internal/ui/ui.go` 第 92-102 行的 switch 语句中，为 `error` 状态添加 case 处理
- [x] 1.2 按任意键时将 `m.state` 设为 `"list"` 并清除 `m.err`

## 2. 验证

- [x] 2.1 编译项目确认无语法错误
