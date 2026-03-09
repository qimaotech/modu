## 1. GitProxy 扩展

- [x] 1.1 在 GitProxy 接口新增 `CheckBranchWorktreeStatus(ctx, repoPath, branch string) (bool, error)` 方法
- [x] 1.2 实现 CheckBranchWorktreeStatus 方法：调用 `git worktree list` 检查分支是否被 worktree 使用

## 2. Engine.CreateWorktree 逻辑修改

- [x] 2.1 修改 module 分支创建逻辑：先检查分支是否存在
- [x] 2.2 分支存在时，调用 CheckBranchWorktreeStatus 检查是否被占用
- [x] 2.3 未被占用时，使用 `git worktree add <path>` 直接 checkout 现有分支（不带 -b 参数）
- [x] 2.4 已被占用时，记录跳过并继续处理下一个 module
- [x] 2.5 主项目保持现有逻辑不变

## 3. 结果输出

- [x] 3.1 跳过 module 时输出 "[SKIP] <module-name>: 分支 <branch> 已被其他 worktree 使用"
- [x] 3.2 创建完成后输出 summary：成功数量和跳过数量

## 4. 测试

- [x] 4.1 编写单元测试：分支存在+未使用 → 应该复用
- [x] 4.2 编写单元测试：分支存在+已使用 → 应该跳过
- [x] 4.3 编写单元测试：分支不存在 → 应该创建新分支
