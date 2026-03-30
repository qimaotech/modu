## 1. 修改 CLI 命令定义

- [x] 1.1 将 `info` 命令的 `Args` 从 `cobra.ExactArgs(1)` 改为 `cobra.MaximumNArgs(1)`

## 2. 实现 feature 推断逻辑

- [x] 2.1 在 `runInfo` 函数中新增 `inferCurrentFeature(cwd, worktreeRoot)` 逻辑
- [x] 2.2 当 `len(args) == 0` 时调用推断逻辑，否则使用 `args[0]`

## 3. 边界处理

- [x] 3.1 当前目录不在 `worktreeRoot` 下时报错：当前目录不在任何 feature 下
- [x] 3.2 当前目录恰好是 `worktreeRoot` 本身时报错：请指定 feature 参数

## 4. 测试验证

- [x] 4.1 在 feature 根目录执行 `modu info` 验证自动推断
- [x] 4.2 在 feature 子目录执行 `modu info` 验证向上回溯
- [x] 4.3 在 `worktreeRoot` 外执行 `modu info` 验证错误提示
- [x] 4.4 执行 `modu info <feature>` 验证显式参数行为不变
- [x] 4.5 执行 `go test ./...` 确保无回归
