## 1. GitProxy 接口变更

- [x] 1.1 在 GitProxy 接口新增 FetchAndSwitchBranch 方法定义

## 2. GitProxy 实现

- [x] 2.1 实现 FetchAndSwitchBranch 方法（fetch + checkout + rebase）
- [x] 2.2 处理分支不存在的情况（从 origin/branch 创建本地分支）

## 3. Engine 修改

- [x] 3.1 修改 UpdateMainProject 使用 FetchAndSwitchBranch 切换到 default-base 分支
- [x] 3.2 修改子模块更新逻辑，使用模块的 base-branch 或全局 default-base

## 4. 测试

- [x] 4.1 Mock 添加 FetchAndSwitchBranchFunc
- [x] 4.2 新增 UpdateMainProject 测试用例
