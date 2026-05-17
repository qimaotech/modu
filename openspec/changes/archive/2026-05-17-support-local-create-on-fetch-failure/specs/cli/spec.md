## ADDED Requirements

### Requirement: CLI create supports explicit local-only mode
`modu create` SHALL support an explicit parameter for feature creation from local branch state without pre-create fetch.

#### Scenario: Local-only parameter skips fetch
- **WHEN** the user runs `modu create <feature> --no-fetch`
- **THEN** the CLI creates the feature from local branch state without pre-create fetch

#### Scenario: Fetch failure keeps original error path with hint
- **WHEN** the user runs `modu create <feature>` and pre-create fetch fails
- **THEN** the CLI reports the original create error
- **AND** the CLI prints a Chinese hint telling the user they can rerun with `--no-fetch` to skip remote fetch and create from local code

#### Scenario: Fetch failure does not prompt in CLI
- **WHEN** the user runs `modu create <feature>` in an interactive terminal and pre-create fetch fails
- **THEN** the CLI does not ask for confirmation and does not continue from local branch state in the same command

#### Scenario: Auto-fetch disabled behaves like local-only mode
- **WHEN** `modu create <feature>` runs with `auto-fetch: false`
- **THEN** the CLI creates the feature from local branch state without pre-create fetch
