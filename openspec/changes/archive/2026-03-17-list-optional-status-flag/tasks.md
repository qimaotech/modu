## 1. 添加 -s flag

- [x] 1.1 在 `cmd/modu/main.go` 的 listCmd 中添加 `-s/--status` flag
- [x] 1.2 在 `runList` 函数中获取 flag 值并传递给 Formatter

## 2. 修改输出格式化

- [x] 2.1 修改 `internal/output/output.go` 的 `FormatListResponse` 方法，增加 `showStatus` 参数
- [x] 2.2 根据 `showStatus` 决定是否显示 `(clean/dirty)` 状态
