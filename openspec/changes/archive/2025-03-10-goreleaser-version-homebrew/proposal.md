## Why

版本信息当前在运行时通过执行 `git describe`/`rev-parse`/`log` 获取，在 CI 或无 Git 环境（如用户安装的二进制）中不可靠或不可用。需要改为在构建时由 GoReleaser 注入版本，并落地 release 规范中的 GoReleaser 打包、GitHub Release 与 Homebrew 接入，同时用 Taskfile 统一发布流程，去掉对 standard-version 的依赖。

## What Changes

- 版本信息（version/commit/date）改为构建时通过 `-ldflags` 注入，移除运行时调用 git 的逻辑。
- 新增 GoReleaser 配置（`.goreleaser.yml`）：多平台构建、GitHub Release、Homebrew tap。
- 发布流程：用 Taskfile 任务完成打 tag、执行 `goreleaser release`，不再使用 standard-version。
- **BREAKING**：移除对 standard-version 的依赖；`install-deps`、`release`、`version:dry-run` 等 task 行为变更。

## Capabilities

### New Capabilities

（无新增能力，仅实现并调整现有 release 能力。）

### Modified Capabilities

- `release`: 版本号改为构建时由 GoReleaser 注入（仍来源于 Git tag）；发布动作为 Taskfile 打 tag + `goreleaser release`，支持 GitHub Release 与 Homebrew tap；不再使用 standard-version。

## Impact

- **代码**：`cmd/modu/main.go` 中删除 `gitDescribe`/`gitCommit`/`gitDate` 及 init 中的调用，保留 version/commit/date 变量供 ldflags 注入。
- **配置**：新增 `.goreleaser.yml`；GitHub 仓库需配置 release 权限；可选单独 Homebrew tap 仓库或使用 GitHub 默认 tap 路径。
- **依赖与工具**：移除 npm/standard-version；发布需安装 GoReleaser（task 可提示或通过 `go install` 安装）。
- **CI/本地**：发布流程统一为 `task release`（或等价）打 tag 后执行 goreleaser。
