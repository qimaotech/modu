## ADDED Requirements

### Requirement: Branch push state inspection
GitProxy SHALL expose an operation that determines whether a local branch is fully contained in a remote branch before deletion.

#### Scenario: Branch has no commits ahead of remote
- **WHEN** the local branch resolves to a remote ref and `remote..branch` has an ahead count of zero
- **THEN** GitProxy reports the branch as pushed

#### Scenario: Branch is ahead of remote
- **WHEN** the local branch resolves to a remote ref and `remote..branch` has an ahead count greater than zero
- **THEN** GitProxy reports the branch as not pushed with the ahead count

#### Scenario: Remote branch cannot be resolved
- **WHEN** the local branch has no upstream and no fallback `origin/<branch>` remote ref
- **THEN** GitProxy reports the branch as not pushed

#### Scenario: Push-state Git command fails
- **WHEN** fetching or push-state inspection fails
- **THEN** GitProxy returns a contextual error and the engine treats the branch as requiring confirmation
