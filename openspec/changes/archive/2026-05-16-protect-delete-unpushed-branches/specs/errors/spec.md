## ADDED Requirements

### Requirement: Unpushed branch deletion error
The system SHALL expose a machine-readable error code for deletion blocked by unpushed local branches.

#### Scenario: Error code for blocked deletion
- **WHEN** deletion is blocked because one or more local branches are not fully pushed
- **THEN** the returned error code is `ERR_UNPUSHED_BRANCH`

#### Scenario: JSON output includes branch details
- **WHEN** JSON output is requested for a delete operation blocked by unpushed branches
- **THEN** the error response includes enough detail to identify the affected repositories, worktrees, branches, and reasons
