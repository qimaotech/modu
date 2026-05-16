## Context

`modu create` already accepts a base branch and `Engine.CreateWorktree(ctx, feature, base)` already applies that value to the main project and to modules without their own `base-branch`. The mismatch is at the entry points: CLI uses a hard-coded `develop` flag default, while TUI silently uses `Config.DefaultBase` and gives users no chance to select a different base for ad hoc second-stage work.

## Goals / Non-Goals

**Goals:**

- Make CLI create default to `.modu.yaml` `default-base` when `--base` is omitted.
- Keep explicit CLI `--base` overrides working.
- Add a TUI base-branch selection/confirmation step between feature-name input and module selection.
- Reuse existing create execution and module `base-branch` override behavior.

**Non-Goals:**

- Do not change gitproxy worktree primitives.
- Do not change `.modu.yaml` schema.
- Do not force a selected base branch to override module-level `base-branch`.
- Do not add remote branch discovery or validation for arbitrary base branch names in this change.

## Decisions

1. Resolve CLI base after loading config.

   The Cobra flag default should be empty, not `develop`. `runCreate` can load config first, then use `eng.Config.DefaultBase` when the flag value is empty. This keeps the default in one source of truth and preserves existing `--base main` behavior.

2. Model TUI base selection as a branch list state.

   Add a `create_base` state with selectable `createBaseOptions`, a list cursor, and a prompt/warning message. It uses `Config.DefaultBase` as the default selected option, reads local and `origin` branches from the workspace repo, and advances to module selection on Enter.

3. Store the selected base on the create flow.

   Keep a `createBase` field on `App` and pass it into `CreateWorktree` during `executeCreateFeature`. This avoids mutating `Config.DefaultBase` for a one-off creation and makes tests able to assert the selected base.

4. Preserve module-specific base override.

   `Engine.CreateWorktree` already chooses `module.BaseBranch` over the provided base for modules. The new TUI/CLI selected base remains the global base input to that existing behavior.

## Risks / Trade-offs

- [Risk] Branch discovery can fail when the workspace repo is missing or git returns an error. -> Mitigation: keep `.modu.yaml` `default-base` as the fallback selectable option and show a Chinese warning.
- [Risk] Users may expect selected base to override module-level `base-branch`. -> Mitigation: document and test the existing precedence instead of changing it implicitly.
- [Risk] Existing CLI scripts relying on implicit `develop` may see different behavior when config uses another default. -> Mitigation: this is the requested behavior and explicit `--base develop` remains available.
