## 1. Git Push-State Primitive

- [x] 1.1 Add branch push-state types and `GitClient` method for checking whether a local branch is fully pushed.
- [x] 1.2 Implement GitProxy push-state inspection with upstream resolution, `origin/<branch>` fallback, ahead-count detection, and contextual errors.
- [x] 1.3 Add gitproxy tests for pushed, ahead, and missing-remote branch states.

## 2. Engine Delete Preflight

- [x] 2.1 Add delete-branch risk types and an engine method that collects unpushed branch candidates for feature deletion.
- [x] 2.2 Add options-based feature deletion that blocks unpushed branches by default and allows them only after explicit confirmation.
- [x] 2.3 Preserve existing dirty-check and branch-slug skip behavior while avoiding partial deletion when unpushed branches exist.
- [x] 2.4 Add engine tests for safe deletion, blocked unpushed deletion, confirmed unpushed deletion, and non-candidate branches.

## 3. CLI Behavior

- [x] 3.1 Add `ERR_UNPUSHED_BRANCH` error code and formatted branch details for blocked deletion.
- [x] 3.2 Update `modu delete` to fail with `ERR_UNPUSHED_BRANCH` when unpushed branches exist.
- [x] 3.3 Add `--allow-unpushed` so users can explicitly confirm deletion without an interactive prompt.

## 4. TUI Behavior

- [x] 4.1 Add TUI state for unpushed-branch risk confirmation with concise branch list rendering.
- [x] 4.2 Update feature deletion flow to preflight after the normal delete confirmation, then delete directly or show the risk confirmation.
- [x] 4.3 Add TUI tests for risk view rendering, confirm-to-delete, and cancel-without-delete behavior.

## 5. Verification

- [x] 5.1 Run gofmt on modified Go files.
- [x] 5.2 Run targeted Go tests for gitproxy, engine, CLI/output, and TUI packages.
- [x] 5.3 Run `go test ./...` if targeted tests pass.
