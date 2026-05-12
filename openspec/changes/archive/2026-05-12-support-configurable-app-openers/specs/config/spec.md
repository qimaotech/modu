## ADDED Requirements

### Requirement: Configurable app openers

配置文件 SHALL support an optional `app-openers` list for declaring extra GUI applications that can open the currently selected project path from the TUI operation menu.

#### Scenario: Load Zed opener example
- **WHEN** `.modu.yaml` contains an `app-openers` entry with `name: zed`, `app: Zed`, `label: Zed`, and `shortcut: z`
- **THEN** config loading succeeds and exposes an app opener whose display text can be rendered as "打开 Zed" with shortcut `z`

#### Scenario: Missing app openers remains compatible
- **WHEN** `.modu.yaml` does not contain `app-openers`
- **THEN** config loading succeeds and the TUI uses only built-in operation menu items

#### Scenario: App opener without shortcut
- **WHEN** an `app-openers` entry omits `shortcut`
- **THEN** config loading succeeds and the opener is available for menu selection without a direct shortcut key

### Requirement: App opener config validation

Each configured app opener MUST include enough information to identify the app and MUST use a valid shortcut shape when a shortcut is provided.

#### Scenario: Missing opener name is invalid
- **WHEN** an `app-openers` entry has an empty `name`
- **THEN** config loading fails with `ERR_CONFIG_INVALID`

#### Scenario: Missing app name is invalid
- **WHEN** an `app-openers` entry has an empty `app`
- **THEN** config loading fails with `ERR_CONFIG_INVALID`

#### Scenario: Shortcut must be a single printable key
- **WHEN** an `app-openers` entry sets `shortcut` to more than one character or to a non-printable key
- **THEN** config loading fails with `ERR_CONFIG_INVALID`

#### Scenario: Shortcut conflict does not invalidate config
- **WHEN** an `app-openers` entry sets `shortcut: c` and `c` is already used by a built-in TUI menu action
- **THEN** config loading still succeeds because shortcut conflict resolution is handled by the TUI menu
