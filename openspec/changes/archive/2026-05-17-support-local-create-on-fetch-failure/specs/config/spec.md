## ADDED Requirements

### Requirement: Auto-fetch controls feature creation fetch
The configuration field `auto-fetch` SHALL control whether feature creation performs pre-create fetch operations before adding worktrees.

#### Scenario: Auto-fetch enabled keeps fetch-first creation
- **WHEN** `.modu.yaml` omits `auto-fetch` or sets `auto-fetch: true`
- **THEN** feature creation attempts to fetch the main project and selected module repositories before creating their worktrees

#### Scenario: Auto-fetch disabled uses local branches
- **WHEN** `.modu.yaml` sets `auto-fetch: false`
- **THEN** feature creation skips pre-create fetch and creates worktrees from local branch state

#### Scenario: Auto-fetch disabled still uses configured base branches
- **WHEN** feature creation runs with `auto-fetch: false`
- **THEN** the main project uses the requested base branch and each module uses its configured `base-branch` or the global `default-base`
