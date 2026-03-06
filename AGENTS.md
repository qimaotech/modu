# AGENTS.md

## 编码

- Go 1.25+，`context.Context` 传参
- 错误用 `errors.New` 或自定义类型
- **语言**: 中文 + 描述性英文变量名

## Git

`feat/fix/refactor/test/docs/chore: <描述>` | 分支: `main` `feature/*` `fix/*` | 提交前 `pre-commit run --all-files`

## 测试

单元 testing+testify，集成 http 测试，覆盖率 80%，命名 `Test<Function>_<Scenario>`

## 工具

- **CI**: `go build/fmt/lint/test ./...`
- **依赖**: Go Modules，`go mod tidy`，不提交 `go.sum`
- **CLI**: Grep 代替 grep，Read 代替 cat，Edit/Write 代替 sed/echo

## 原则

- 不读取 docs/archive 文件夹
- 不生成敏感信息（密码/密钥/token）
- 不猜测 URL
- 简洁响应，不改未请求代码
- scope 外先确认，用 auto memory 记录经验
- DRY / KISS / YAGNI
