## 1. 版本注入与 main.go

- [x] 1.1 移除 `cmd/modu/main.go` 中 `gitDescribe`、`gitCommit`、`gitDate` 函数及 `init()` 内对三者的调用，保留 `version`、`commit`、`date` 变量（默认值 "dev"/"unknown"）供 ldflags 注入
- [x] 1.2 确认 `modu version` 子命令仅读取上述变量并输出，无运行时 git 调用

## 2. GoReleaser 配置

- [x] 2.1 在项目根目录新增 `.goreleaser.yml`：配置 builds（Darwin/Linux，arm64/amd64）、archives、GitHub release
- [x] 2.2 在 `.goreleaser.yml` 中配置 Homebrew tap（Formula 名称与仓库或 tap 路径）

## 3. Taskfile 发布流程

- [x] 3.1 删除 Taskfile 中 standard-version 相关 task：`install-deps`、`release`（旧）、`version:dry-run`
- [x] 3.2 新增 `release` task：校验当前分支为 main 或 master，执行 `goreleaser release`（若未安装则提示安装方式）
- [x] 3.3 新增或保留“查看下一版本/打 tag”的说明或 task（如 `version:next` 仅输出建议的 tag，或文档说明 `git tag vx.y.z` 后执行 `task release`）

## 4. 文档与收尾

- [x] 4.1 在 README 或 release 文档中说明：正式版本需通过 goreleaser 构建、如何安装 goreleaser、发布步骤（打 tag + task release）、本地 `go build` 得到版本为 "dev"
