package errors

import "errors"

// 错误码定义
var (
	ErrConfigInvalid    = errors.New("ERR_CONFIG_INVALID")
	ErrGitExec          = errors.New("ERR_GIT_EXEC")
	ErrDirtyWorktree    = errors.New("ERR_DIRTY_WORKTREE")
	ErrPartialFailure   = errors.New("ERR_PARTIAL_FAILURE")
	ErrFeatureExists    = errors.New("ERR_FEATURE_EXISTS")
	ErrFeatureNotFound  = errors.New("ERR_FEATURE_NOT_FOUND")
	ErrModuleNotFound   = errors.New("ERR_MODULE_NOT_FOUND")
	ErrInvalidOperation = errors.New("ERR_INVALID_OPERATION")
)

// Code 返回错误码字符串
func Code(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrConfigInvalid) {
		return "ERR_CONFIG_INVALID"
	}
	if errors.Is(err, ErrGitExec) {
		return "ERR_GIT_EXEC"
	}
	if errors.Is(err, ErrDirtyWorktree) {
		return "ERR_DIRTY_WORKTREE"
	}
	if errors.Is(err, ErrPartialFailure) {
		return "ERR_PARTIAL_FAILURE"
	}
	if errors.Is(err, ErrFeatureExists) {
		return "ERR_FEATURE_EXISTS"
	}
	if errors.Is(err, ErrFeatureNotFound) {
		return "ERR_FEATURE_NOT_FOUND"
	}
	if errors.Is(err, ErrModuleNotFound) {
		return "ERR_MODULE_NOT_FOUND"
	}
	if errors.Is(err, ErrInvalidOperation) {
		return "ERR_INVALID_OPERATION"
	}
	return "ERR_UNKNOWN"
}
