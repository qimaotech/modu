## 1. 修改配置向导

- [x] 1.1 修改 `config_wizard.go` 中 `doSaveConfig()` 方法，将配置文件保存路径从当前目录改为 workspace 目录
- [x] 1.2 在 `doSaveConfig()` 中添加 workspace 目录创建逻辑（如果不存在）
- [x] 1.3 在 `doSaveConfig()` 中添加 worktree-root 目录创建逻辑（如果不存在）

## 2. 实现 Git 仓库初始化

- [x] 2.1 新增 `ensureGitRepo()` 函数，检查 workspace 是否为 Git 仓库
- [x] 2.2 在 `ensureGitRepo()` 中实现 `git init` 调用
- [x] 2.3 在 `ensureGitRepo()` 中实现 `git checkout -b <base>` 调用
- [x] 2.4 在 `doSaveConfig()` 中调用 `ensureGitRepo()`

## 3. 错误处理

- [x] 3.1 处理 Git 操作失败情况，返回明确错误信息
- [x] 3.2 在 `doSaveConfig()` 中正确传递错误到上层

## 4. 测试验证

- [x] 4.1 测试 workspace 目录不存在场景
- [x] 4.2 测试 workspace 已是 Git 仓库场景
- [x] 4.3 测试 workspace 是普通目录（非 Git）场景
- [x] 4.4 测试 Git 命令执行失败场景
