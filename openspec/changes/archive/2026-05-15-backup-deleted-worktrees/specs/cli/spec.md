## ADDED Requirements

### Requirement: CLI runs delete backup cleanup after config load
CLI commands that successfully load an existing `.modu.yaml` via normal or scan config loading SHALL invoke delete backup cleanup once before executing their main command behavior. Cleanup failure SHALL be reported as a warning on stderr and MUST NOT block the requested command or corrupt JSON stdout.

#### Scenario: Configured command starts
- **WHEN** `modu create`, `modu delete`, `modu default-select`, `modu list`, `modu info`, `modu status`, `modu update`, `modu init`, or `modu config scan` successfully loads existing configuration
- **THEN** the command invokes delete backup cleanup once before its main behavior

#### Scenario: Cleanup warning does not block command
- **WHEN** startup cleanup fails after configuration loading succeeds
- **THEN** the CLI prints a warning to stderr and continues with the requested command

#### Scenario: Cleanup warning preserves JSON stdout
- **WHEN** startup cleanup fails and the requested command uses `-o json`
- **THEN** the cleanup warning is not written to stdout

#### Scenario: Command without existing config load does not clean
- **WHEN** `modu version`, help output, or `modu config create` runs without loading an existing config
- **THEN** delete backup cleanup is not required to run

### Requirement: CLI delete exposes backup path
`modu delete <feature>` SHALL report the created backup archive path when deletion succeeds.

#### Scenario: Text delete output includes backup path
- **WHEN** `modu delete my-feature` succeeds with text output
- **THEN** the output includes the feature name and backup archive path

#### Scenario: JSON delete output includes backup path
- **WHEN** `modu delete my-feature -o json` succeeds
- **THEN** the JSON response includes `backupPath`
