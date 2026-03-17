## 1. 修改全局样式定义

- [x] 1.1 修改 `headerStyle`，使用 `AdaptiveColor` 定义深色（236/86）和浅色（254/25）配色
- [x] 1.2 修改 `itemStyle`，使用 `AdaptiveColor` 定义深色（252）和浅色（245）配色
- [x] 1.3 修改 `selectedItemStyle`，使用 `AdaptiveColor` 定义深色（86）和浅色（21）配色
- [x] 1.4 修改 `errorStyle`，使用 `AdaptiveColor` 定义深色/浅色配色（均为 196）
- [x] 1.5 修改 `successStyle`，使用 `AdaptiveColor` 定义深色（82）和浅色（28）配色

## 2. 验证测试

- [x] 2.1 运行单元测试验证无回归 `go test ./...`
- [x] 2.2 在深色终端背景下启动 TUI，验证显示正常
- [x] 2.3 在浅色终端背景下启动 TUI，验证显示正常
