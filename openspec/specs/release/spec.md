# modu 分发与升级规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

定义构建、分发与升级方式。

## 构建与分发

- 使用 **GoReleaser** 打包。
- 支持 **brew install** 接入。
- 支持平台：Darwin / Linux，arm64 / amd64。
- 一键安装脚本：`curl -sSL https://get.modu.sh | sh`（若提供）。

## 版本与升级

- 版本号来自 Git（如 `git describe --tags --abbrev=0`），通过 `modu version` 展示（含 commit、date）。
- 内置 **modu upgrade** 命令（可选）：通过 GitHub API 检查新版本并替换本地二进制。

## 与代码的对应

- 版本：`cmd/modu/main.go` 中 version/commit/date 变量及 `version` 子命令；GoReleaser 配置（若存在）在项目根目录。
