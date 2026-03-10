## Context

- 当前 `cmd/modu/main.go` 在 `init()` 中通过 `git describe` / `rev-parse` / `log` 获取 version/commit/date，在无 Git 或 CI 环境中不可靠。
- release 规范已约定使用 GoReleaser 与 brew，但项目尚未有 `.goreleaser.yml`，且发布流程依赖 npm standard-version。
- 需要统一为：构建时注入版本、GoReleaser 负责打包与发布、Taskfile 负责打 tag 与调用 goreleaser。

## Goals / Non-Goals

**Goals:**

- 版本信息在构建时通过 ldflags 注入，不依赖运行时 Git。
- 使用 GoReleaser 完成多平台构建、GitHub Release 与 Homebrew tap 发布。
- 发布流程由 Taskfile 定义（打 tag + `goreleaser release`），移除 standard-version。

**Non-Goals:**

- 不实现一键安装脚本（get.modu.sh）或 modu upgrade 命令；仅保证版本可展示与分发方式就绪。

## Decisions

1. **版本注入方式**：使用 `-ldflags -X main.version=... -X main.commit=... -X main.date=...`，由 GoReleaser 在 build 阶段注入。理由：与 GoReleaser 原生支持一致，无需运行时依赖；替代方案（运行时读环境变量）不利于未通过 goreleaser 的本地 build，故不采用。
2. **GoReleaser 配置**：单文件 `.goreleaser.yml` 于项目根目录，包含 builds（多平台）、archives、release（GitHub）、homebrew tap。Homebrew 使用 GitHub 默认 tap 路径（如 `owner/homebrew-tap` 或仓库内 Formula）由配置指定。
3. **发布流程**：Taskfile 提供 `release` task：校验当前分支（main/master）、可选版本 bump 或由用户先打 tag，然后执行 `goreleaser release`。不再提供 `install-deps`（standard-version）与 `version:dry-run`（standard-version）；可提供 `task version:next` 或文档说明如何打 tag。
4. **GoReleaser 安装**：文档与 task 描述中建议 `go install github.com/goreleaser/goreleaser/v2@latest` 或使用官方 install script；不在 Taskfile 内强制安装。

## Risks / Trade-offs

- **Risk**：未通过 GoReleaser 的本地 `go build` 得到的二进制 version 为默认值（如 "dev"）。**Mitigation**：README 说明正式版本需通过 `goreleaser build` 或 CI 构建；本地开发可用 `task build` 或显式 ldflags。
- **Risk**：首次配置 Homebrew tap 需 GitHub 仓库或 token 权限。**Mitigation**：在 design/README 中列出所需权限与可选 tap 仓库创建步骤。
- **Trade-off**：移除 standard-version 后，版本号需手动或脚本打 tag；若需自动 bump，可后续在 Taskfile 中加简单脚本（如基于上一 tag 的 patch bump），本次不纳入。
