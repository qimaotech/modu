# modu

多模块 Git Worktree 管理工具，用于简化多仓库协作开发流程。

## 功能特性

| 命令     | 说明                             |
| -------- | -------------------------------- |
| `create` | 创建 feature 工作树              |
| `delete` | 删除 feature 工作树              |
| `list`   | 列出所有 feature 工作树          |
| `info`   | 查看 feature 详情                |
| `init`   | 初始化仓库（克隆所有配置的仓库） |
| `status` | 显示所有模块的脏状态             |
| `update` | 更新代码（fetch + rebase）       |
| `config` | 配置管理（创建/扫描）           |
| `tui`    | 启动交互式 TUI 界面              |
| `version`| 显示版本信息                     |

## 安装

### 安装脚本（推荐，可指定版本）

```bash
# 安装指定版本（推荐，便于复现与确认）
curl -sSfL https://github.com/qimaotech/modu/-/raw/main/install.sh | sh -s v1.0.0

# 安装最新
curl -sSfL https://github.com/qimaotech/modu/-/raw/main/install.sh | sh -s
```

### go install

```bash
# 指定版本
go install github.com/qimaotech/modu/cmd/modu@v1.0.0
# 或最新
go install github.com/qimaotech/modu/cmd/modu@latest
```

> 注意：确保 `$(go env GOPATH)/bin` 或 `$(go env GOBIN)` 在你的 PATH 中。

## 使用方法

### 配置文件

创建 `.modu.yaml` 配置文件：

```yaml
workspace: ~/workspace/main
worktree-root: ~/workspace
default-base: develop
concurrency: 5
auto-fetch: true
strict-dirty-check: true

modules:
  - name: frontend
    url: https://github.com/example/frontend.git
  - name: backend
    url: https://github.com/example/backend.git
```

**配置项说明：**

| 字段                 | 必填 | 说明                      |
| -------------------- | ---- | ------------------------- |
| `workspace`          | 是   | 裸仓库/主仓库所在目录     |
| `worktree-root`      | 是   | 特性分支代码存放目录      |
| `default-base`       | 是   | 默认基准分支 (如 develop) |
| `concurrency`        | 否   | 并发数，默认 5            |
| `auto-fetch`         | 否   | 操作前自动 fetch          |
| `strict-dirty-check` | 否   | 删除前强制脏检查          |
| `modules`            | 是   | 模块列表                  |

**模块配置：**

| 字段          | 必填 | 说明             |
| ------------- | ---- | ---------------- |
| `name`        | 是   | 模块名称         |
| `url`         | 是   | Git 仓库地址     |
| `base-branch` | 否   | 覆盖全局默认分支 |

### 自动 .gitignore 更新

执行 `modu init` 或 `modu config scan` 时会自动更新主仓库的 `.gitignore` 文件，添加所有模块目录，避免模块代码被意外提交到主仓库。

### 命令示例

```bash
# 初始化所有仓库
modu init
modu init --scan  # 自动扫描并添加模块

# 创建 feature 分支
modu create my-feature
modu create my-feature --base main
modu create my-feature --modules frontend,backend  # 只创建指定模块

# 列出所有 worktree
modu list
modu list -v  # 显示详细信息（模块、分支、状态）

# 查看详情
modu info my-feature

# 删除 worktree
modu delete my-feature
modu delete my-feature --force  # 强制删除（跳过脏检查）

# 查看脏状态
modu status

# 指定配置文件
modu list -c /path/to/config.yaml

# JSON 格式输出
modu list -o json

# 启动 TUI
modu
modu tui

# 更新代码（主项目或 feature）
modu update                    # 更新主项目（workspace + 所有模块）
modu update my-feature         # 更新指定 feature 的 worktree

# 配置管理
modu config create                           # 交互式创建配置文件
modu config create --module "frontend=..."   # 指定模块创建配置
modu config scan                             # 扫描目录自动发现模块
modu config scan --module "backend=..."       # 扫描并添加模块

# 查看版本信息
modu version

### TUI 快捷键

| 按键   | 说明                   |
| ------ | ---------------------- |
| ↑/↓    | 上下选择 feature       |
| Enter  | 确认删除选中 feature   |
| o      | 用 VS Code 打开主项目 |
| q/esc  | 退出 TUI              |

# 创建配置文件
modu config create
modu config create --module "frontend=git@github.com:example/frontend.git"
modu config create --scan  # 自动扫描并添加模

# 扫描目录添加模块
modu config scan
```

## 开发

```bash
# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 运行单元测试
go test -v ./internal/...

# 运行 E2E 测试
go test -v -tags=e2e -run TestE2E .

# 代码检查
golangci-lint run

# 本地构建（版本显示为 dev/unknown，仅开发用）
go build -o modu ./cmd/modu
# 或
task build
```

### 构建与发布

- **版本来源**：正式发布的二进制版本号由 [GoReleaser](https://goreleaser.com/) 在构建时通过 ldflags 注入；本地 `go build` 未注入，`modu version` 会显示 `dev` / `unknown`。
- **安装 GoReleaser**：`go install github.com/goreleaser/goreleaser/v2@latest`（或参见 [官方安装说明](https://goreleaser.com/install)）。
- **本地多平台构建（不发布）**：`goreleaser build --snapshot --clean`，产物在 `dist/`。
- **发布新版本**（需在 main/master 分支）：
  1. 打 tag：`git tag v1.2.3`
  2. 执行：`task release`（内部会调用 `goreleaser release`）
  3. 需配置 `GITHUB_TOKEN`（或所用 SCM 的 token）；若使用 Homebrew tap，需创建 tap 仓库并配置写权限，见 `.goreleaser.yml` 中 `brews` 及环境变量说明。
- 查看当前 tag 与发布提示：`task version:next`。
- **详细步骤**（GitHub Release + Homebrew tap）：见 [docs/RELEASE.md](docs/RELEASE.md)。

## 技术栈

- Go 1.25+
- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI 框架
