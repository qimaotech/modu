## 1. AddModule 逻辑修改

- [x] 1.1 在 AddModule 函数中添加分支存在检查逻辑
- [x] 1.2 分支存在时调用 CheckBranchWorktreeStatus 检查是否被占用
- [x] 1.3 未被占用时调用 CreateWorktreeFromExistingBranch 复用分支
- [x] 1.4 已被占用时输出跳过提示并返回成功

## 2. 删除 feature 时跳过非 module 目录

- [x] 2.1 修改 DeleteWorktree 函数：只删除配置中存在的 module 目录
- [x] 2.2 对于不在配置中的目录，只删除目录文件，不尝试删除 git 分支

## 3. 测试

- [x] 3.1 编写单元测试：AddModule 分支存在+未使用 → 应该复用
- [x] 3.2 编写单元测试：AddModule 分支存在+已使用 → 应该跳过
- [x] 3.3 编写单元测试：AddModule 分支不存在 → 应该创建新分支
