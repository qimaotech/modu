## 1. 依赖添加

- [x] 1.1 添加 `github.com/atotto/clipboard` 依赖: `go get github.com/atotto/clipboard`

## 2. 列表视图复制功能

- [x] 2.1 在 `handleListKey` 中添加 `case "c"` 处理，复制选中项路径并显示临时消息

## 3. 菜单视图复制功能

- [x] 3.1 在 `renderMenu` 中添加 "复制路径 (c)" 菜单项（位置 1）
- [x] 3.2 在 `handleMenuKey` 中调整 `menuLen` 从 `2/4` 改为 `3/5`
- [x] 3.3 在 `handleMenuKey` 中添加 `case "c"` 处理，复制路径并关闭菜单
- [x] 3.4 在 `handleMenuKey` 的 `enter` 分支中，将原有 index 调整（1→2, 2→3, 3→4），插入复制路径的 enter 处理

## 4. 错误处理

- [x] 4.1 当 `env.MainProject` 为 nil 时，显示错误消息并进入 error state

## 5. 测试

- [x] 5.1 添加 `TestCopyPath` 单元测试覆盖列表视图和菜单视图的复制场景
