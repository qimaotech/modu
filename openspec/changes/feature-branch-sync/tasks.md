## 1. GitProxy 层

- [x] 1.1 在 `internal/gitproxy/gitproxy.go` 的 `GitClient` 接口中添加 `RemoteBranchExists(ctx context.Context, repoURL, branch string) bool`
- [x] 1.2 在 `internal/gitproxy/gitproxy_impl.go` 中使用 `git ls-remote --heads` 实现 `RemoteBranchExists`
- [x] 1.3 为 `RemoteBranchExists` 添加单元测试，覆盖：分支存在、分支不存在、仓库不存在、网络错误

## 2. Engine 层

- [x] 2.1 在 `Engine` 中添加 `GetModulesWithRemoteBranch(ctx context.Context, branch string) (map[string]bool, error)` 方法
- [x] 2.2 使用 errgroup 实现并发查询，限制并发数为 `Config.Concurrency`
- [x] 2.3 为 `GetModulesWithRemoteBranch` 添加单元测试，覆盖：全部有分支、全部无分支、部分有分支、URL为空

## 3. UI 层

- [x] 3.1 在 `internal/ui/ui.go` 中更新 `SelectModules` 签名，添加 `remoteHasBranch map[string]bool` 参数
- [x] 3.2 更新 `NewModuleSelector` 以接收并使用 `remoteHasBranch` 进行预选逻辑
- [x] 3.3 更新代码库中所有 `SelectModules` 的调用者，传入新参数

## 4. CLI 集成

- [x] 4.1 在 `modu create` 命令中，调用 `eng.GetModulesWithRemoteBranch(ctx, feature)` 然后再调用 `ui.SelectModules`
- [x] 4.2 将结果传递给 `ui.SelectModules` 用于预选
- [x] 4.3 优雅处理错误：查询失败时使用空 map 并记录警告

## 5. 验证

- [x] 5.1 运行 `go build ./...` 验证编译
- [x] 5.2 运行 `go test ./...` 确保所有测试通过
- [x] 5.3 手动测试：在交互模式下运行 `modu create feature/test` 并验证预选行为
