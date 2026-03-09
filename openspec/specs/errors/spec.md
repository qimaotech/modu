# modu 错误处理规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

统一错误码、错误包装、日志规范及 `-o json` 机器输出协议。

## 错误码

| 错误码 | 说明 |
|--------|------|
| `ERR_CONFIG_INVALID` | modu.yaml 格式或路径非法、必填缺失 |
| `ERR_GIT_EXEC` | Git 命令执行失败（需包含 exit code / stderr 上下文） |
| `ERR_DIRTY_WORKTREE` | 脏检查拦截，存在未提交修改 |
| `ERR_PARTIAL_FAILURE` | 并发操作中部分成功、部分失败 |
| `ERR_FEATURE_EXISTS` | Feature 目录已存在 |
| `ERR_FEATURE_NOT_FOUND` | Feature 目录不存在 |
| `ERR_MODULE_NOT_FOUND` | 模块路径不存在（如 git status 时目录缺失） |
| `ERR_INVALID_OPERATION` | 非法操作（预留） |

## 错误包装

- gitproxy 层抛出的错误须包含上下文，例如：`fmt.Errorf("[%s] git worktree add failed: %w", moduleName, err)`。
- 上层可再次 wrap，保留链式原因，不截断。

## 错误日志

- **上下文**：当前命令与参数、配置文件路径、workspace/worktree-root、涉及模块名。
- **链式原因**：多步操作保留每一步失败原因。
- **外部命令**：Git 调用的完整 stderr/stdout 写入日志。
- **路径与状态**：失败时的绝对路径、目标分支、当前分支等。

## 并发错误聚合

- 使用 `errors.Join` 或等价方式等待当前批次结束。
- 终端：打印“成功 X 个，失败 Y 个”及失败模块的具体 Git 报错。
- `-o json`：`errors` 数组包含所有失败项的上下文。

## 机器输出协议（-o json）

### 成功响应示例

```json
{
  "success": true,
  "action": "create",
  "feature": "feature-login",
  "results": [
    { "module": "auth-svc", "status": "success", "path": "/worktrees/feature-login/auth-svc" }
  ],
  "errors": []
}
```

### 失败响应示例

```json
{
  "code": "ERR_DIRTY_WORKTREE",
  "message": "cannot delete: uncommitted changes detected",
  "data": {
    "feature": "feat-login",
    "dirty_modules": [
      { "name": "auth-api", "files": ["main.go", "config.yaml"] }
    ]
  }
}
```

## CLI 与 TUI 行为矩阵

| 场景 | CLI 行为 | TUI 行为 |
|------|----------|----------|
| 配置/路径错误 | 直接退出并打印错误 | 界面提示，不退出 |
| Git 失败 | 带出完整错误信息 | 界面提示并允许重试 |
| 并发部分失败 | 打印成功/失败汇总 | 显示失败模块列表 |
| 脏检查失败 | 退出码非 0 | 弹窗阻止删除 |

## 与代码的对应

- 实现：`internal/errors`（错误码定义、Code()）；`internal/output` 与 cmd 层负责 JSON 序列化与退出码。
