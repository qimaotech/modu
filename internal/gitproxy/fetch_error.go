package gitproxy

import stderrors "errors"

// FetchError 表示创建或更新前的 git fetch 阶段失败。
type FetchError struct {
	RepoPath string
	Err      error
}

// Error 返回底层 fetch 错误描述。
func (e *FetchError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// Unwrap 返回底层错误，保留 errors.Is / errors.As 能力。
func (e *FetchError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// IsFetchError 判断错误是否来自 git fetch 阶段。
func IsFetchError(err error) bool {
	var fetchErr *FetchError
	return stderrors.As(err, &fetchErr)
}
