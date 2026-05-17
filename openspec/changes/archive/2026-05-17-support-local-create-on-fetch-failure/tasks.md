## 1. Gitproxy Fetch Control

- [x] 1.1 Add a typed way to identify pre-create fetch failures while preserving `ERR_GIT_EXEC` wrapping and contextual stderr.
- [x] 1.2 Add gitproxy worktree creation paths that skip fetch for new branches and existing local branches.
- [x] 1.3 Add gitproxy tests for fetch failure identification, no-fetch new branch creation, and no-fetch existing branch reuse.

## 2. Engine Create Policy

- [x] 2.1 Add create options or equivalent engine input for fetch-enabled versus local-only creation while preserving existing `CreateWorktree` callers.
- [x] 2.2 Apply `Config.AutoFetch` during main project and module creation so `auto-fetch: false` skips pre-create fetch.
- [x] 2.3 Return distinguishable pre-create fetch failures when fetch-enabled creation fails before `git worktree add`.
- [x] 2.4 Preserve rollback behavior when local-only creation fails after partially creating worktrees.
- [x] 2.5 Add engine tests for fetch-enabled failure, local-only creation, auto-fetch disabled, and rollback after local-only failure.

## 3. CLI Create Interaction

- [x] 3.1 Add a `--no-fetch` flag to `modu create` that creates from local branch state without pre-create fetch.
- [x] 3.2 When fetch-enabled CLI create fails during pre-create fetch, keep the existing error output and append a Chinese `--no-fetch` hint.
- [x] 3.3 Ensure CLI create never prompts or retries local creation in the same command after fetch failure.
- [x] 3.4 Add CLI tests for `--no-fetch`, fetch failure hint output, no prompt/retry behavior, and auto-fetch disabled behavior.

## 4. TUI Create Interaction

- [x] 4.1 Add a TUI confirmation state/message that displays the concrete fetch error and asks whether to retry from local code.
- [x] 4.2 Retry the pending create operation without fetch when the user confirms.
- [x] 4.3 Route declined retry or cancellation through the existing creation error feedback path.
- [x] 4.4 Add TUI tests for error+prompt display, accept retry, decline/cancel, and auto-fetch disabled prompt avoidance.

## 5. Documentation And Verification

- [x] 5.1 Update README configuration/usage notes for `auto-fetch` and local fallback behavior.
- [x] 5.2 Run `go fmt ./...`.
- [x] 5.3 Run `go test ./...`.
- [x] 5.4 Run `pre-commit run --all-files`.
