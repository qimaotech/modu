## Context

当前 TUI 在显示错误页面时，state 切换为 "error"，但按键处理只覆盖了 list、menu、modules、confirm 四种状态。当 state 为 error 时，按任何键都不会触发处理逻辑，导致程序卡死。

## Goals / Non-Goals

**Goals:**
- 让用户在 error 页面按任意键可以返回到 list 状态

**Non-Goals:**
- 不修改错误展示样式
- 不修改其他状态的行为

## Decisions

在 `ui.go:92-102` 的 switch 语句中，为 `error` 状态添加 case：

```go
case "error":
    m.state = "list"
    m.err = nil
```

这是最简单的实现，按任意键直接返回 list 状态并清除错误。

## Risks / Trade-offs

无显著风险，这是用户体验的改进。
