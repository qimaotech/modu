## ADDED Requirements

### Requirement: TUI create can retry locally after fetch failure
The TUI SHALL show the concrete fetch error and allow users to retry feature creation from local branch state when pre-create fetch fails during feature creation.

#### Scenario: Show fetch error with local retry confirmation
- **WHEN** the user confirms project/module selection and pre-create fetch fails
- **THEN** the TUI shows the concrete fetch error message
- **AND** the TUI shows a Chinese confirmation prompt asking whether to create from local code and warning that remote branch conflicts may occur later

#### Scenario: User accepts local retry
- **WHEN** the TUI local retry prompt is visible
- **AND** the user confirms continuing locally
- **THEN** the TUI retries the pending feature creation without pre-create fetch and returns to the list view on success

#### Scenario: User declines local retry
- **WHEN** the TUI local retry prompt is visible
- **AND** the user cancels or declines
- **THEN** the TUI stops creation and shows the original fetch failure through the existing error feedback path

#### Scenario: Auto-fetch disabled avoids retry prompt
- **WHEN** the loaded config has `auto-fetch: false`
- **THEN** TUI feature creation skips pre-create fetch and does not show the local retry prompt
