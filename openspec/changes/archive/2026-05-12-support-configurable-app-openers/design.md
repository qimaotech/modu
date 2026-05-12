## Context

当前 TUI 操作菜单在 `internal/ui/ui.go` 中直接维护固定菜单项，并分别用硬编码分支处理 VS Code、Codex、复制路径、更新代码、Modules 管理和删除。配置加载由 `internal/config` 读取 `.modu.yaml`，目前没有描述额外打开工具的字段。

用户希望通过配置扩展操作菜单中的打开工具，例如 Zed、Cursor。该配置需要在 TUI 启动时读取，并根据本机是否安装对应 app 决定是否展示；如果快捷键与已有菜单项或其他 app opener 冲突，则菜单项仍可通过上下选择和 Enter 执行，但不展示、不绑定冲突快捷键。

用户描述中的 `.mudu.yaml` 按项目现状理解为现有默认配置文件 `.modu.yaml`，避免引入第二套配置入口。

## Goals / Non-Goals

**Goals:**

- 在 `.modu.yaml` 中支持可选的自定义 app opener 列表。
- TUI 启动时解析并过滤已安装的 app opener。
- 操作菜单展示已安装 app opener，按配置顺序放在内置打开项之后、复制/更新/管理/删除等非打开操作之前。
- 唯一且未被内置菜单占用的快捷键可触发对应 app opener；冲突快捷键不展示也不绑定。
- 保持未配置时的现有行为。

**Non-Goals:**

- 不支持用户在 TUI 内编辑 app opener 配置。
- 不新增跨平台 GUI app 发现依赖；优先使用系统已有命令完成安装检测。
- 不改变 worktree 路径选择规则，仍打开当前选中条目的主项目路径。
- 不把未安装 app 作为错误展示给用户。

## Decisions

1. **配置字段使用 `app-openers`**

   增加 `Config.AppOpeners []AppOpener`，YAML 字段为 `app-openers`。每项包含稳定名称 `name`、系统 app 名称 `app`、可选展示名 `label`、可选快捷键 `shortcut`。

   示例：

   ```yaml
   app-openers:
     - name: zed
       app: Zed
       label: Zed
       shortcut: z
     - name: cursor
       app: Cursor
       label: Cursor
       shortcut: c
   ```

   `label` 只表示 UI 中 "打开 <label>" 的显示文本；为空时使用 `app`。`shortcut` 为空时该项仅支持菜单选择后 Enter 执行。

2. **配置校验只验证结构，不验证安装状态**

   `name` 与 `app` 必填；`shortcut` 如存在必须是单个可打印字符。是否安装属于运行环境状态，由 TUI 启动时判断，避免同一配置在不同机器上因为安装状态不同而无法加载。

3. **安装检测放在 TUI 层**

   TUI 初始化时从 `m.Engine.Config.AppOpeners` 构建运行时菜单项。macOS 使用 `open -Ra <app>` 检测 app 是否可解析；其他平台可先返回不可用或后续补充平台实现。检测失败不进入错误状态，只是不展示该 opener。

4. **菜单项统一建模**

   用内部 `menuItem` 结构描述 label、shortcut、action、enabled path 需求等信息，渲染和执行都基于同一个菜单数组。这样可以消除当前 render 与 enter/key 分支各自维护顺序造成的偏移风险，也方便插入配置化 opener。

5. **快捷键冲突在运行时降级**

   TUI 先保留内置快捷键集合（例如 `o/x/c/u/m/d/q/esc`），再统计可见 app opener 的配置快捷键。快捷键与内置键冲突，或同一个键被多个可见 opener 使用时，对应 opener 渲染为 "打开 <label>"，不显示 `(key)`，也不响应该 key；用户仍可选中后按 Enter 打开。

## Risks / Trade-offs

- [Risk] macOS `open -Ra` 对 app 名称大小写或别名解析依赖系统行为 → Mitigation: 文档说明 `app` 应填写 `open -a` 可识别的应用名称，并用单元测试覆盖命令构造与失败隐藏逻辑。
- [Risk] 继续在多个分支维护菜单顺序容易产生索引错位 → Mitigation: 重构为单一菜单构建函数，渲染、上下导航、Enter 执行和快捷键处理共享同一菜单数组。
- [Risk] 用户为 Cursor 配置 `c` 时会与复制路径冲突 → Mitigation: 不报错、不绑定 `c`，菜单仍显示 "打开 Cursor"，用户可以 Enter 执行。
- [Risk] 非 macOS 平台暂时无法可靠检测 GUI app → Mitigation: 将检测逻辑包在可替换函数中；当前实现无新增依赖，后续可补充平台实现。
