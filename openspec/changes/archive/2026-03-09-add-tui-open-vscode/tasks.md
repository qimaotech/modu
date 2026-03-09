## 1. TUI 打开 VS Code 功能实现

- [x] 1.1 在 `handleListKey` 函数中添加 `case "o":` 分支处理
- [x] 1.2 实现获取选中 feature 的主项目路径逻辑
- [x] 1.3 使用 `exec.Command("code", path).Start()` 异步打开 VS Code
- [x] 1.4 添加主项目为空的错误处理

## 2. UI 优化

- [x] 2.1 在列表视图底部 hints 中添加 `[o] 打开项目` 提示
