## 1. Config Schema

- [x] 1.1 Add `AppOpener` config model and `Config.AppOpeners` YAML field mapped to `app-openers`.
- [x] 1.2 Validate configured openers: `name` and `app` are required, optional `shortcut` is a single printable key, and shortcut conflicts remain valid for TUI resolution.
- [x] 1.3 Add config unit tests for the Zed example, missing `app-openers`, shortcut omission, missing required fields, invalid shortcut shape, and a built-in shortcut conflict.

## 2. TUI Menu Model

- [x] 2.1 Introduce a shared operation-menu item model used by rendering, cursor bounds, Enter execution, and direct shortcut execution.
- [x] 2.2 Build menu items from built-in actions plus installed configured app openers, preserving config order and inserting custom openers after built-in open actions.
- [x] 2.3 Add app availability detection for configured openers, using `open -Ra <app>` on macOS and hiding unavailable openers without surfacing an error.
- [x] 2.4 Resolve shortcut conflicts against built-in operation-menu keys and other visible configured openers; conflicting openers remain selectable but render without a shortcut and do not bind that key.
- [x] 2.5 Implement configured app opener execution with the current selected main project path and the configured app name.

## 3. Documentation

- [x] 3.1 Update README configuration documentation with an `app-openers` example using Zed and shortcut `z`.
- [x] 3.2 Update TUI operation-menu shortcut documentation to describe configured app openers and conflict behavior.

## 4. Verification

- [x] 4.1 Add TUI tests for installed Zed rendering as "打开 Zed (z)", missing app hiding, Cursor `c` conflict rendering without `(c)`, unique shortcut execution, Enter execution, and missing-path error handling.
- [x] 4.2 Run `gofmt` on changed Go files.
- [x] 4.3 Run focused config and UI tests.
- [x] 4.4 Run the project lint/test/build checks required before completion where practical.
