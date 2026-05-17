## Context

`modu create` currently relies on gitproxy methods that always run `git fetch` before `git worktree add`. This keeps branch information fresh, but it also means a temporary network failure blocks feature creation even when the local base branch is available.

The configuration already contains `auto-fetch`, but create flows do not use it. CLI and TUI also have different interaction shapes: CLI should preserve its error-first behavior and offer an explicit rerun parameter, while TUI can ask for confirmation in place.

## Goals / Non-Goals

**Goals:**

- Add an explicit CLI parameter for local-only feature creation.
- Keep CLI fetch failure behavior compatible: print the original error and add a local-only create hint.
- Let TUI users continue feature creation from local branches after seeing the concrete fetch error.
- Make `auto-fetch: false` skip pre-create fetch for main project and selected modules.
- Preserve existing rollback behavior and branch reuse semantics.
- Keep CLI local-only hints and TUI local retry prompts in Chinese.

**Non-Goals:**

- Do not change `modu update`; update remains fetch/rebase oriented.
- Do not make CLI auto-confirm or retry local creation after fetch failure.
- Do not introduce network reachability probes outside normal git commands.
- Do not change branch naming or worktree directory rules.

## Decisions

### Decision: Use explicit local-only create for CLI

CLI should keep the original failure behavior when fetch fails. It should report the same create error path and append a Chinese hint such as `可使用 --no-fetch 跳过远程拉取，基于本地代码创建`. Users who want local-only creation rerun `modu create <feature> --no-fetch`.

Alternative considered: prompt interactively in CLI after fetch failure. This changes the existing command flow and is less script-friendly.

### Decision: Let TUI confirm local retry in place

TUI should show the specific fetch error and then ask whether to retry creation from local code. If the user confirms, TUI retries the pending create operation with fetch disabled. If the user declines, it returns to the existing error feedback path.

Alternative considered: require TUI users to leave the TUI and rerun CLI with a flag. That is awkward because the TUI already has the pending feature name and module selection.

### Decision: Model fetch failure separately from worktree creation failure

Gitproxy should preserve contextual `ERR_GIT_EXEC` errors and expose enough information for the engine/CLI/TUI to distinguish a pre-create fetch failure from `git worktree add` failure. When the caller explicitly disables fetch, worktree creation should run without a fetch attempt.

Alternative considered: parse error strings in CLI/TUI. That is brittle because git stderr varies by transport and locale.

### Decision: Use existing `auto-fetch` as the default policy switch

When `auto-fetch` is true or omitted, create attempts fetch first. When false, create skips pre-create fetch and uses local branch state directly. The CLI `--no-fetch` parameter overrides the config for a single create command.

Alternative considered: add a new `offline-create` config field. That adds another knob for a behavior already described by `auto-fetch`.

### Decision: Keep remote branch preselection best-effort

Remote branch detection for module preselection remains best-effort. If it cannot query the remote, selection still opens with default-selected modules only; the TUI local retry prompt is reserved for actual create-time fetch failure.

Alternative considered: prompt during selection when remote query fails. That would interrupt users before they have committed to creating anything.

## Risks / Trade-offs

- Remote branch conflict risk -> warn users that local-only creation cannot verify remote branch state and a later push may conflict.
- Stale base branch risk -> require explicit CLI `--no-fetch`, `auto-fetch: false`, or TUI confirmation before creating without fetch.
- Interface churn risk -> keep old gitproxy methods where practical and add option/no-fetch variants rather than forcing unrelated callers to change.
- Partial creation risk -> rely on existing engine rollback if any local-only worktree creation fails.
