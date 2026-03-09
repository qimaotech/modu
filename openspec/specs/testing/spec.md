# modu 测试规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

规定覆盖率目标、关键测试场景与 E2E 要求。

## 覆盖率目标

- **internal/engine**：高逻辑覆盖（使用 Mock Git），建议 100% 关键路径。
- **整体项目**：强制 **> 85%**；新增代码需同步补测，覆盖率不降低。

## 关键测试矩阵

| 模块 | 测试场景 | 预期行为 |
|------|----------|----------|
| Config | 缺失 workspace 字段 | 报错 `ERR_CONFIG_INVALID` |
| Engine | 并发创建时某一模块失败 | 触发回滚，删除已创建的 worktree 目录 |
| Engine | 在 Dirty 目录下执行 delete | 拦截操作，返回该模块名 / ERR_DIRTY_WORKTREE |
| GitProxy | 解析 git status 输出 | 正确识别 M、??、D 为 Dirty |
| E2E | init → create → delete 完整流程 | 物理目录与 `git worktree list` 一致 |

## 脏检查测试

- 临时目录修改文件不提交，断言 `modu delete` 返回 `ERR_DIRTY_WORKTREE`。
- 添加 Untracked 文件，验证脏检查识别。
- 使用 `--force` 验证跳过脏检查。

## 并发与隔离

- 并发数为 1 时行为与串行一致。
- 多模块同时写入同一 worktree-root 时，路径隔离正确（各 feature/模块独立目录）。

## CLI E2E

每个子命令至少一条 E2E：

- `modu config scan`、`modu config scan --export`（若支持）
- `modu init`（可用 fixture 或 mock 远端）
- `modu create` / `modu list` / `modu info` / `modu delete`
- 断言退出码、stdout/stderr 关键内容。

## 与代码的对应

- 实现：`internal/config`、`internal/engine/engine_test.go`、`internal/gitproxy` 测试、`internal/output/output_test.go`；E2E 位于对应测试包或集成测试目录。
