package errors

import (
	"errors"
	"testing"
)

func TestCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantCode: "",
		},
		{
			name:     "ErrConfigInvalid",
			err:      ErrConfigInvalid,
			wantCode: "ERR_CONFIG_INVALID",
		},
		{
			name:     "ErrGitExec",
			err:      ErrGitExec,
			wantCode: "ERR_GIT_EXEC",
		},
		{
			name:     "ErrDirtyWorktree",
			err:      ErrDirtyWorktree,
			wantCode: "ERR_DIRTY_WORKTREE",
		},
		{
			name:     "ErrPartialFailure",
			err:      ErrPartialFailure,
			wantCode: "ERR_PARTIAL_FAILURE",
		},
		{
			name:     "ErrFeatureExists",
			err:      ErrFeatureExists,
			wantCode: "ERR_FEATURE_EXISTS",
		},
		{
			name:     "ErrFeatureNotFound",
			err:      ErrFeatureNotFound,
			wantCode: "ERR_FEATURE_NOT_FOUND",
		},
		{
			name:     "ErrModuleNotFound",
			err:      ErrModuleNotFound,
			wantCode: "ERR_MODULE_NOT_FOUND",
		},
		{
			name:     "ErrInvalidOperation",
			err:      ErrInvalidOperation,
			wantCode: "ERR_INVALID_OPERATION",
		},
		{
			name:     "wrapped error",
			err:      errors.New("wrapped: " + ErrConfigInvalid.Error()),
			wantCode: "ERR_UNKNOWN",
		},
		{
			name:     "unknown error",
			err:      errors.New("some random error"),
			wantCode: "ERR_UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Code(tt.err)
			if got != tt.wantCode {
				t.Errorf("Code(%v) = %q, want %q", tt.err, got, tt.wantCode)
			}
		})
	}
}
