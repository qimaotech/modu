# modu 分发与升级规范（delta）

**变更**: goreleaser-version-homebrew | **基规范**: openspec/specs/release/spec.md

## MODIFIED Requirements

### Requirement: 版本信息由构建时注入并展示

版本号（version）、commit 短哈希、构建日期（date）MUST 在构建时通过 ldflags 注入到 `cmd/modu/main.go` 的变量中，并由 GoReleaser 在发布构建中设置；版本号来源 SHALL 为 Git tag（如 `v1.2.3`）。`modu version` SHALL 输出上述 version、commit、date，且不依赖运行时执行 git 命令。

#### Scenario: 通过 GoReleaser 构建后查看版本
- **WHEN** 使用 GoReleaser 完成构建并运行生成的二进制 `modu version`
- **THEN** 输出包含与当前 tag 一致的版本号、对应 commit、构建时间

#### Scenario: 未注入时的默认值
- **WHEN** 使用普通 `go build` 且未传入 ldflags
- **THEN** `modu version` 输出中 version/commit/date 为默认占位值（如 "dev"/"unknown"），不报错

### Requirement: 使用 GoReleaser 打包并发布到 GitHub 与 Homebrew

项目 MUST 在根目录提供 `.goreleaser.yml`。GoReleaser SHALL 负责：多平台二进制构建（Darwin/Linux，arm64/amd64）、生成 GitHub Release 并上传制品、生成并提交 Homebrew Formula（支持 `brew install` 接入）。发布动作 SHALL 由维护者在本地或 CI 执行 `goreleaser release`（或通过 Taskfile 封装的 task）触发，且仅在 main/master 分支允许发布。

#### Scenario: 执行发布后 GitHub Release 存在
- **WHEN** 在 main 或 master 分支对当前 commit 打 tag 并执行 goreleaser release
- **THEN** 在 GitHub 仓库的 Releases 中生成对应 tag 的 release，并包含各平台二进制

#### Scenario: Homebrew 可安装
- **WHEN** 用户配置的 Homebrew tap 已更新（或使用项目提供的 tap 仓库）
- **THEN** 执行 `brew install <tap>/modu` 可安装与 GitHub Release 一致的二进制

### Requirement: 发布流程由 Taskfile 定义且不依赖 standard-version

发布流程 SHALL 由 Taskfile 提供 task（如 `release`）：校验当前分支为 main 或 master、指导或执行打 tag、调用 `goreleaser release`。项目 MUST NOT 依赖 npm 或 standard-version 做版本号或 changelog 管理；版本 bump 与打 tag 由维护者手动或通过 Taskfile 内简单脚本完成。

#### Scenario: 在正确分支执行 release task
- **WHEN** 当前分支为 main 或 master 且已存在指向当前 commit 的 tag，执行 `task release`（或等价命令）
- **THEN** 调用 goreleaser 完成构建并发布到 GitHub 与 Homebrew，无 standard-version 步骤

#### Scenario: 在错误分支执行 release task
- **WHEN** 当前分支非 main 且非 master 时执行 release task
- **THEN** task 报错并退出，不执行 goreleaser
