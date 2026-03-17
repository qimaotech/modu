## 1. 目录名转换函数实现

- [x] 1.1 新增 `featureToDirName` 函数（feature/hello → feature-hello）
- [x] 1.2 新增 `dirNameToFeature` 函数（feature-hello → feature/hello）

## 2. 修改各操作函数

- [x] 2.1 修改 CreateWorktree 使用 featureToDirName
- [x] 2.2 修改 ListWorktrees 使用 dirNameToFeature
- [x] 2.3 修改 DeleteWorktree 使用 featureToDirName
- [x] 2.4 修改 UpdateWorktree 使用 featureToDirName
- [x] 2.5 修改 GetWorktreeInfo 使用 featureToDirName
- [x] 2.6 修改 AddModule 使用 featureToDirName
- [x] 2.7 修改 RemoveModule 使用 featureToDirName

## 3. 测试验证

- [x] 3.1 单元测试通过
- [x] 3.2 验证扁平结构 features 仍然正常工作
- [x] 3.3 手动测试 `modu list` 正确显示带 / 的 feature
